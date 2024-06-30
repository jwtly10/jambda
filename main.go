package main

import (
	"context"
	"log/slog"
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
)

// @title Jambda - Serverless framework
// @version 0.1
// @description A WIP serverless framework for running functions similar to AWS Lambda

// @contact.name jwtly10/Jambda
// @contact.url https://www.github.com/jwtly10/jambda

// @host localhost:8080
// @BasePath /v1/api
func main() {
	logger := logging.NewLogger(false, slog.LevelDebug.Level())

	fs := afero.NewOsFs()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration:", err)
	}

	db, err := db.ConnectDB(cfg)
	if err != nil {
		logger.Fatal("Database connection failed:", err)
	}
	logger.Info("Database connected")

	// Init Router
	router := api.NewAppRouter(logger)
	router.SetupSwagger()

	// Setup services
	functionRepo := repository.NewFunctionRepository(db)
	fileService := service.NewFileService(functionRepo, logger, fs)

	dockerService := service.NewDockerService(logger, *functionRepo)

	// Setup middlewares
	loggerMw := &middleware.RequestLoggerMiddleware{Log: logger}
	dockerMw := &middleware.DockerMiddleware{Log: logger, Ds: *dockerService}

	// Setup routes

	// File routes
	fileHandler := handlers.NewFileHandler(logger, *fileService)
	routes.NewFileRoutes(router, logger, *fileHandler, loggerMw)

	// Gateway routes
	gatewayService := service.NewGatewayService(logger)
	gatewayHandler := handlers.NewGatewayHandler(logger, *gatewayService)
	routes.NewGatewayRoutes(router, logger, *gatewayHandler, loggerMw, dockerMw)

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
