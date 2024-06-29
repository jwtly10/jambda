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
	_ "github.com/jwtly10/jambda/docs"
	"github.com/jwtly10/jambda/pkg/logging"
)

// @title Jambda - Serverless framework
// @version 0.1
// @description A WIP serverless framework for running functions similar to AWS Lambda

// @contact.name jwtly10
// @contact.url https://www.github.com/jwtly10/jambda

// @host localhost:8080
// @BasePath /v1
func main() {
	logger := logging.NewLogger(false, slog.LevelDebug.Level())

	uploadHandler := handlers.NewFunctionHandler(logger)

	router := api.NewAppRouter(logger)
	router.SetupSwagger()

	loggerMw := &middleware.RequestLoggerMiddleware{Log: logger}

	routes.NewUploadRoutes(router, logger, *uploadHandler, loggerMw)

	logger.Info("Server starting on port 8080...")

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
