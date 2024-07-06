import { useEffect, useState, useRef } from 'react'
import { getFunctions, deleteFunction, putFunction, postFunction } from './api/clientApi';
import { Container, Row, Col, ListGroup, Button, Modal, Form, Navbar} from 'react-bootstrap';
import Terminal from 'react-console-emulator'


function App() {

    // Functions state
    const [functions, setFunctions] = useState([]);
    const [selectedFunction, setSelectedFunction] = useState(null);
    const [newFunction, setNewFunction] = useState({
        name: '',
        file: null,
        config: {
            trigger: '',
            image: '',
            type: '',
            port: 8080,
            envVars: {}
        }
    });

    const commands = {
        echo: {
            description: 'Echo a passed string.',
            usage: 'echo <string>',
            fn: (...args) => args.join(' ')
        }
    }

    // Modal states
    const [showEditModal, setShowEditModal] = useState(false);
    const [showNewModal, setShowNewModal] = useState(false);
    const [showConfirmModal, setShowConfirmModal] = useState(false);

    // Validation states
    const [editValidated, setEditValidated] = useState(false);
    const [newValidated, setNewValidated] = useState(false);


    // Error message states
    const [newErrorMessage, setNewErrorMessage] = useState("")
    const [editErrorMessage, setEditErrorMessage] = useState("")

    // On mount
    useEffect(() => {
        (async () => {
            await getActiveFuncs()
        })();
        return () => { }
    }, [])

    // Reset the terminal when a new function is selected
    useEffect(() => {
        if (terminalRef.current) {
            terminalRef.current.clearStdout();
        }
    }, [selectedFunction]);

    const getActiveFuncs = async () => {
        try {
            const res = await getFunctions()
            if (res && Array.isArray(res)) {
                const sortedFunctions = res.sort((a, b) => b.id - a.id);
                setFunctions(sortedFunctions);
            } else {
                setFunctions([]);
            }
        } catch (e) {
            console.error(e)
        }
    }

    const handleConfigChange = (type, field, value) => {
        switch (type) {
            case "new":
                setNewFunction(prev => ({
                    ...prev,
                    config: {
                        ...prev.config,
                        [field]: value
                    }
                }));
                break;
            case "edit":
                setSelectedFunction(prev => ({
                    ...prev,
                    configuration: {
                        ...prev.configuration,
                        [field]: value
                    }
                }));
                break
        }
    }

    const handleNewSubmit = async (event) => {
        const form = event.currentTarget;
        event.preventDefault();
        if (form.checkValidity() === false) {
            event.stopPropagation();
        } else {
            try {
                await postFunction(newFunction.file, JSON.stringify(newFunction.config), newFunction.name)
                setShowNewModal(false);
                await getActiveFuncs()
            } catch (e) {
                console.error(e)
                setNewErrorMessage(e.response.data.message)
            }
        }
        setNewValidated(true);
    };

    const handleEditSubmit = async (event) => {
        const form = event.currentTarget;
        event.preventDefault();
        if (form.checkValidity() === false) {
            event.stopPropagation();
        } else {
            try {
                await putFunction(selectedFunction.external_id, JSON.stringify(selectedFunction.configuration), selectedFunction.name)
                setShowEditModal(false);
                await getActiveFuncs()
            } catch (e) {
                console.error(e)
                setEditErrorMessage(e.response.data.message)
            }
        }
        setEditValidated(true);
    };

    const handleSelectFunction = (func) => {
        console.log(func)
        setSelectedFunction(func);
    };

    const handleEditOpen = async () => {
        setShowEditModal(true);
    };

    const handleEditClose = async () => {
        setShowEditModal(false);
        await getActiveFuncs()
    };

    const handleNewOpen = async () => {
        setShowNewModal(true);
    };

    const handleNewClose = async () => {
        setShowNewModal(false);
        await getActiveFuncs()
    };

    const handleDeleteConfirm = () => {
        setShowConfirmModal(true);
    };

    const handleDeleteClose = () => {
        setShowConfirmModal(false);
    };

    const handleDelete = async () => {
        setFunctions(functions.filter(f => f.id !== selectedFunction.id));
        try {
            await deleteFunction(selectedFunction.external_id)
        } catch (e) {
            console.error(e)
        }
        setSelectedFunction(null);
        setShowConfirmModal(false);
    };

    const terminalRef = useRef(null);

    const fetchLogs = (id) => {
        terminalRef.current.pushToStdout(`Initializing log stream for function '${id}' ...`);
        // TODO. Implement chunk streaming endpoint
        // to stream logs based on 
        const mockLogs = [
            "Log entry 1: Process started",
            "Log entry 2: Performing operation",
            "Log entry 3: Performing Saving data",
        ];

        mockLogs.forEach((log, index) => {
            setTimeout(() => {
                terminalRef.current.pushToStdout(log);
            }, 2000 * (index + 1)); 
        });

    };

    return (
        <>
            <Navbar bg="dark" variant="dark" className="mb-4">
                <Container>
                    <Navbar.Brand >Jambda Function Configuration</Navbar.Brand>
                </Container>
            </Navbar>
            <Container>
                <Row>
                    <Col md={4}>
                        <Button variant="primary" className="mb-3" onClick={handleNewOpen}>New Function</Button>
                        <h2>Functions</h2>
                        <ListGroup>
                            {functions.map(func => (
                                <ListGroup.Item key={func.id} action onClick={() => handleSelectFunction(func)}>
                                    <b>{func.name}</b> [{func.external_id}]
                                </ListGroup.Item>
                            ))}
                        </ListGroup>
                    </Col>
                    <Col md={8}>
                        {selectedFunction && (
                            <Row>
                                <Col md={4}>
                                    <div className="mb-2">
                                        <h4>{selectedFunction.name}</h4>
                                        <p>External Id: {selectedFunction.external_id}</p>
                                        <p>Trigger: {selectedFunction.configuration.trigger}</p>
                                        <p>Type: {selectedFunction.configuration.type}</p>
                                        <p>Image: {selectedFunction.configuration.image}</p>
                                        <p>Port: {selectedFunction.configuration.port}</p>
                                        <Button variant="primary" onClick={handleEditOpen}>Edit</Button>{' '}
                                        <Button variant="danger" onClick={handleDeleteConfirm}>Delete</Button>
                                    </div>
                                </Col>
                                <Col md={8}>
                                    <Button variant="secondary" className="mb-2" onClick={() => fetchLogs(selectedFunction.external_id)}>Stream Logs</Button>
                                    <Terminal
                                        commands={{}}
                                        ref={terminalRef}
                                        readOnly={true}
                                        welcomeMessage={'Log output:'}
                                        promptLabel=''
                                        style={{
                                            height: '400px',
                                            width: '100%',
                                            backgroundColor: '#242424',
                                            color: 'lime',
                                        }}
                                    />
                                </Col>
                            </Row>
                        )}
                    </Col>
                </Row>

                <Modal show={showEditModal} onHide={handleEditClose}>
                    <Modal.Header closeButton>
                        <Modal.Title>Edit Function</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        {selectedFunction && (
                            <Form noValidate validated={editValidated} onSubmit={handleEditSubmit}>
                                <Form.Group className="mb-3" controlId="formEditName">
                                    <Form.Label>Function Name</Form.Label>
                                    <Form.Control
                                        required
                                        type="text"
                                        placeholder="Enter name"
                                        value={selectedFunction.name}
                                        onChange={e => setSelectedFunction({ ...selectedFunction, name: e.target.value })}
                                    />
                                    <Form.Control.Feedback type="invalid">
                                        Please provide a function name.
                                    </Form.Control.Feedback>
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formEditTrigger">
                                    <Form.Label>Trigger</Form.Label>
                                    <Form.Select
                                        required
                                        value={selectedFunction.configuration.trigger}
                                        onChange={e => handleConfigChange('edit', 'trigger', e.target.value)}
                                    >
                                        <option value="">Select Trigger</option>
                                        <option value="http">HTTP</option>
                                        <option value="cron">CRON</option>
                                    </Form.Select>
                                    <Form.Control.Feedback type="invalid">
                                        Please select a trigger.
                                    </Form.Control.Feedback>
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formEditImage">
                                    <Form.Label>Image</Form.Label>
                                    <Form.Select
                                        required
                                        value={selectedFunction.configuration.image}
                                        onChange={e => handleConfigChange('edit', 'image', e.target.value)}
                                    >
                                        <option value="">Select Image</option>
                                        <option value="golang:1.22">golang:1.22</option>
                                        <option value="openjdk:17-jdk">openjdk:17-jdk</option>
                                        <option value="openjdk:21-jdk">openjdk:21-jdk</option>
                                    </Form.Select>
                                    <Form.Control.Feedback type="invalid">
                                        Please select an image.
                                    </Form.Control.Feedback>
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formEditType">
                                    <Form.Label>Type</Form.Label>
                                    <Form.Select
                                        required
                                        value={selectedFunction.configuration.type}
                                        onChange={e => handleConfigChange('edit', 'type', e.target.value)}
                                    >
                                        <option value="">Select Type</option>
                                        <option value="REST">REST</option>
                                        <option value="SINGLE">SINGLE</option>
                                    </Form.Select>
                                    <Form.Control.Feedback type="invalid">
                                        Please select a type.
                                    </Form.Control.Feedback>
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formEditPort">
                                    <Form.Label>Port</Form.Label>
                                    <Form.Control
                                        required
                                        type="number"
                                        placeholder="Enter port number"
                                        value={selectedFunction.configuration.port}
                                        onChange={e => handleConfigChange('edit', 'port', e.target.value)}
                                    />
                                    <Form.Control.Feedback type="invalid">
                                        Please enter a port number.
                                    </Form.Control.Feedback>
                                </Form.Group>
                            </Form>
                        )}
                    </Modal.Body>
                    <Modal.Footer>
                        {editErrorMessage !== "" && (
                            <p className="text-danger">{editErrorMessage}</p>
                        )}
                        <Button variant="secondary" onClick={handleEditClose}>
                            Close
                        </Button>
                        <Button variant="primary" onClick={handleEditSubmit}>
                            Save Changes
                        </Button>
                    </Modal.Footer>
                </Modal>



                <Modal show={showNewModal} onHide={handleNewClose}>
                    <Modal.Header closeButton>
                        <Modal.Title>New Function</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <Form noValidate validated={newValidated} onSubmit={handleNewSubmit}>
                            <Form.Group className="mb-3" controlId="formNewName">
                                <Form.Label>Function Name</Form.Label>
                                <Form.Control
                                    required
                                    type="text"
                                    placeholder="Enter name"
                                    value={newFunction.name}
                                    onChange={e => setNewFunction({ ...newFunction, name: e.target.value })}
                                />
                                <Form.Control.Feedback type="invalid">
                                    Please provide a function name.
                                </Form.Control.Feedback>
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formFile">
                                <Form.Label>Function Binary</Form.Label>
                                <Form.Control
                                    required
                                    type="file"
                                    onChange={e => setNewFunction({ ...newFunction, file: e.target.files[0] })}
                                />
                                <Form.Control.Feedback type="invalid">
                                    Please upload a function binary.
                                </Form.Control.Feedback>
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formTrigger">
                                <Form.Label>Trigger</Form.Label>
                                <Form.Select
                                    required
                                    value={newFunction.config.trigger}
                                    onChange={e => handleConfigChange('new', 'trigger', e.target.value)}
                                >
                                    <option value="">Select Trigger</option>
                                    <option value="http">HTTP</option>
                                    <option value="cron">CRON</option>
                                </Form.Select>
                                <Form.Control.Feedback type="invalid">
                                    Please select a trigger.
                                </Form.Control.Feedback>
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formImage">
                                <Form.Label>Image</Form.Label>
                                <Form.Select
                                    required
                                    value={newFunction.config.image}
                                    onChange={e => handleConfigChange('new', 'image', e.target.value)}
                                >
                                    <option value="">Select Image</option>
                                    <option value="golang:1.22">golang:1.22</option>
                                    <option value="openjdk:17-jdk">openjdk:17-jdk</option>
                                    <option value="openjdk:21-jdk">openjdk:21-jdk</option>
                                </Form.Select>
                                <Form.Control.Feedback type="invalid">
                                    Please select an image.
                                </Form.Control.Feedback>
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formType">
                                <Form.Label>Type</Form.Label>
                                <Form.Select
                                    required
                                    value={newFunction.config.type}
                                    onChange={e => handleConfigChange('new', 'type', e.target.value)}
                                >
                                    <option value="">Select Type</option>
                                    <option value="REST">REST</option>
                                    <option value="SINGLE">SINGLE</option>
                                </Form.Select>
                                <Form.Control.Feedback type="invalid">
                                    Please select a type.
                                </Form.Control.Feedback>
                            </Form.Group>
                            <Form.Group className="mb-3" controlId="formPort">
                                <Form.Label>Port</Form.Label>
                                <Form.Control
                                    required
                                    type="number"
                                    placeholder="Enter port number"
                                    value={newFunction.config.port}
                                    onChange={e => handleConfigChange('new', 'port', e.target.value)}
                                />
                                <Form.Control.Feedback type="invalid">
                                    Please enter a port number.
                                </Form.Control.Feedback>
                            </Form.Group>
                            <Modal.Footer>
                                {newErrorMessage != "" && (
                                    <p className="text-danger">{newErrorMessage}</p>
                                )}
                                <Button variant="secondary" onClick={handleNewClose}>
                                    Close
                                </Button>
                                <Button type="submit">Save</Button>
                            </Modal.Footer>
                        </Form>
                    </Modal.Body>
                </Modal>




                <Modal show={showConfirmModal} onHide={handleDeleteClose} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>Confirm Delete</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>Are you sure you want to delete this function?</Modal.Body>
                    <Modal.Footer>
                        <Button variant="secondary" onClick={handleDeleteClose}>Cancel</Button>
                        <Button variant="danger" onClick={handleDelete}>Delete</Button>
                    </Modal.Footer>
                </Modal>
            </Container>
        </>
    );
}

export default App;

