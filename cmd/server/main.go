package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
)

func main() {
	// Load environment variables from .env file (optional in Docker)
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connection
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	db, close := database.NewWithHost(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
		host,
	)
	defer close()

	// Tune connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %s", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	// Initialize repositories
	prodRepo := models.NewProductsRepository(db)
	catRepo := models.NewCategoriesRepository(db)

	// Initialize services
	catalogService := catalog.NewCatalogService(prodRepo)

	// Initialize handlers
	catalogHandler := catalog.NewCatalogHandler(catalogService)
	categoriesHandler := categories.NewCategoriesHandler(catRepo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", catalogHandler.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", catalogHandler.HandleGetByCode)
	mux.HandleFunc("GET /categories", categoriesHandler.HandleGet)
	mux.HandleFunc("GET /categories/{code}", categoriesHandler.HandleGetByCode)
	mux.HandleFunc("POST /categories", categoriesHandler.HandlePost)

	// Set up the HTTP server
	// Listen on 0.0.0.0 to accept connections from outside the container
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8484"
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: mux,
	}

	// Start the server
	go func() {
		log.Printf("Starting server on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}

		log.Println("Server stopped gracefully")
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	stop()
}
