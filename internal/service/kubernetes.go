package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/logging"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type KubernetesService struct {
	log     logging.Logger
	kcli    *kubernetes.Clientset
	kconfig *rest.Config
}

func NewKubernetesService(log logging.Logger) *KubernetesService {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("failed to create kubernetes config", err)
	}

	kcli, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create kubernetes client", err)
	}

	return &KubernetesService{
		log:     log,
		kcli:    kcli,
		kconfig: config,
	}
}

func (ks *KubernetesService) DeployToMinikube(ctx context.Context, functionId string, deploymentName string, namespace string, config *data.FunctionConfig) error {

	replicas := int32(3)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deploymentName,
							Image: "local/app:" + functionId,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(*config.Port),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the deployment
	deploymentsClient := ks.kcli.AppsV1().Deployments(namespace)
	ks.log.Info("Creating deployment...")
	result, err := deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	ks.log.Infof("Created deployment %q.\n", result.GetObjectMeta().GetName())

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": deploymentName,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,                           // Public port
					TargetPort: intstr.FromInt(*config.Port), // Internal container port
					NodePort:   0,                            // Dynmically assign host port
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	servicesClient := ks.kcli.CoreV1().Services(namespace)
	ks.log.Info("Creating service...")
	resultService, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	ks.log.Infof("Created service %q.\n", resultService.GetObjectMeta().GetName())

	time.Sleep(2 * time.Second)

	ks.log.Info("Setup port forwarding...")
	var dynamicPort int
	for _, port := range resultService.Spec.Ports {
		ks.log.Infof("NodePort for %s on port %d is %d", resultService.Name, port.Port, port.NodePort)
		dynamicPort = int(port.NodePort)
	}

	stopChan := make(chan struct{})
	ks.startPortForward(resultService.Name, "default", dynamicPort, 80, stopChan)

	ks.log.Infof("Successfully setup port forwarding for k8 service")

	return nil
}

func (ks *KubernetesService) startPortForward(serviceName string, namespace string, localPort int, remotePort int, stopChan <-chan struct{}) {
	cmd := exec.Command("kubectl", "port-forward", "svc/"+serviceName, fmt.Sprintf("%d:%d", localPort, remotePort), "-n", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	go func() {
		if err := cmd.Start(); err != nil {
			ks.log.Fatalf("Failed to start port forwarding: %v", err)
		}

		ks.log.Infof("Port forwarding started for %s on ports %d -> %d", serviceName, localPort, remotePort)

		err := cmd.Wait()
		if err != nil {
			ks.log.Errorf("Port forwarding stopped for %s: %v", serviceName, err)
		}
	}()

	go func() {
		<-stopChan
		if err := cmd.Process.Kill(); err != nil {
			ks.log.Errorf("Failed to kill port forwarding process: %v", err)
		}
		ks.log.Info("Port forwarding terminated")
	}()

	ks.log.Infof("Port forwarding started")
}

func (ks *KubernetesService) setupPortForwarding(config *rest.Config, serviceName string, localPort int) {
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	go func() {
		podName, err := ks.getFirstPodName(config, "default", serviceName)
		if err != nil {
			ks.log.Fatalf("Failed to get pod name for service %s: %v", serviceName, err)
		}

		ks.log.Infof("Found pod name: %s", podName)

		ks.forwardPorts(config, podName, "default", []string{fmt.Sprintf("%d:80", localPort)}, stopChan, readyChan)
	}()

	<-readyChan
	fmt.Printf("Port forwarding for service %s set up successfully on port %d\n", serviceName, localPort)
}

func (ks *KubernetesService) getFirstPodName(config *rest.Config, namespace, serviceName string) (string, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", "my-app-"+serviceName),
	})
	if err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for service %s", serviceName)
	}

	return pods.Items[0].Name, nil
}

func (ks *KubernetesService) forwardPorts(config *rest.Config, podName, namespace string, ports []string, stopChan, readyChan chan struct{}) {
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		ks.log.Fatalf("Failed to create round tripper: %v", err)
	}

	hostIP := strings.TrimPrefix(config.Host, "https://")

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	ks.log.Infof("Forward path: %s", path)

	url := &url.URL{Scheme: "https", Path: path, Host: hostIP}
	ks.log.Infof("URL: %s", url.String())

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST",
		url)

	ks.log.Infof("Dialer: %v", dialer)

	pf, err := portforward.New(
		dialer,
		ports,
		stopChan,
		readyChan,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		ks.log.Fatalf("Failed to create port forwarder: %v", err)
	}

	if err := pf.ForwardPorts(); err != nil {
		ks.log.Fatalf("Failed to forward ports: %v", err)
	}
}
