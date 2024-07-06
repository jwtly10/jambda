package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jwtly10/jambda/api"
	"github.com/jwtly10/jambda/api/handlers"
	"github.com/jwtly10/jambda/api/middleware"
	"github.com/jwtly10/jambda/api/routes"
	"github.com/jwtly10/jambda/config"
	_ "github.com/jwtly10/jambda/docs"
	"github.com/jwtly10/jambda/internal/db"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/jwtly10/jambda/internal/repository"
	"github.com/jwtly10/jambda/internal/service"
	"github.com/spf13/afero"
	"go.uber.org/zap/zapcore"
)

// @title Jambda - Serverless framework
// @version 0.1
// @description A WIP serverless framework for running functions similar to AWS Lambda

// @contact.name jwtly10/Jambda
// @contact.url https://www.github.com/jwtly10/jambda

// @host localhost:8080
// @BasePath /v1/api
func main() {
	logger := logging.NewLogger(false, zapcore.DebugLevel)

	fs := afero.NewOsFs()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration:", err)
		panic("Unable to load config")
	}

	db, err := db.ConnectDB(cfg)
	if err != nil {
		logger.Fatal("Database connection failed:", err)
		panic("Unable to connect to database")
	}
	logger.Info("Database connected")

	// Init Router
	router := api.NewAppRouter(logger)

	// Setup global middleware
	loggerMw := &middleware.RequestLoggerMiddleware{Log: logger}
	router.Use(loggerMw)

	router.SetupSwagger()
	router.ServeStaticFiles("./jambda-frontend/dist")

	// Setup services
	configValidator := service.NewConfigValidator(logger)
	functionRepo := repository.NewFunctionRepository(db)

	fileService := service.NewFileService(functionRepo, logger, fs, *configValidator)
	gatewayService := service.NewGatewayService(logger)
	functionService := service.NewFunctionService(functionRepo, logger, *fileService, *configValidator)

	dockerService := service.NewDockerService(logger, *functionRepo)

	// Setup specific middlewares
	dockerMw := &middleware.DockerMiddleware{Log: logger, Ds: *dockerService}
	usageMw := &middleware.UsageMiddleware{Log: logger}

	// Setup routes

	// File routes
	fileHandler := handlers.NewFunctionHandler(logger, *functionService)
	routes.NewFunctionRoutes(router, logger, *fileHandler)

	// Gateway routes
	gatewayHandler := handlers.NewGatewayHandler(logger, *gatewayService)
	routes.NewGatewayRoutes(router, logger, *gatewayHandler, dockerMw, usageMw)

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		logger.Info("Starting server on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting server", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down server", err)
	}

	logger.Info("Server gracefully stopped")
}
