package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"payment-gateway/internal/config"
	"payment-gateway/internal/routes"
	"payment-gateway/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	router *chi.Mux
	db     *mongo.Database
}

// NewApp initializes a new application instance
func NewApp() (*App, error) {
	// Load environment variables
	config.LoadEnv()

	// Connect to the database
	db, err := utils.ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize router and setup routes
	router := chi.NewRouter()
	routes.SetupRoutes(router, db)

	return &App{
		router: router,
		db:     db,
	}, nil
}

// Start runs the HTTP server with graceful shutdown
func (a *App) Start(ctx context.Context) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: a.router,
	}

	// Log server start
	log.Println("Server running on port", port)

	// Start the server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
	}()

	// Graceful shutdown handling
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Println("Shutting down server...")
		return server.Shutdown(context.Background())
	}
}
