package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"weather-app/internal/api"
	"weather-app/internal/platform/database"
	"weather-app/internal/platform/email"
	"weather-app/internal/platform/scheduler"
	"weather-app/internal/platform/weatherprovider"
	"weather-app/internal/service"
)

func main() {
	// System configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appBaseURL := os.Getenv("APP_BASE_URL")
	if appBaseURL == "" {
		appBaseURL = fmt.Sprintf("http://localhost:%s", port) // For local development
	}

	weatherAPIKey := os.Getenv("WEATHERAPI_COM_KEY")
	if weatherAPIKey == "" {
		log.Fatal("Error: WEATHERAPI_COM_KEY environment variable not set.")
	}

	// Database Configuration
	dbCfg := database.DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
	if dbCfg.Host == "" {
		dbCfg.Host = "localhost"
	}
	if dbCfg.Port == "" {
		dbCfg.Port = "5432"
	}
	if dbCfg.User == "" {
		dbCfg.User = "myuser"
	}
	if dbCfg.Password == "" {
		dbCfg.Password = "mysecretpassword"
	}
	if dbCfg.DBName == "" {
		dbCfg.DBName = "weatherappdb"
	}
	if dbCfg.SSLMode == "" {
		dbCfg.SSLMode = "disable"
	} // For local development

	// Database Connection & Migrations
	db, err := database.ConnectDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Construct DB URL for migrations
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName, dbCfg.SSLMode)

	migrationsDir, _ := filepath.Abs("./migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Printf("Trying to find migrations at: %s", migrationsDir)
	}

	if err := database.RunMigrations(dbURL, migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Dependencies
	// Weather Provider
	weatherClient := weatherprovider.NewClient(weatherAPIKey)

	// Repositories
	subRepo := database.NewPGSubscriptionRepository(db)

	// Email Service (Placeholder)
	emailService := email.NewLogEmailService()

	// Business Logic Services
	subscriptionSvc := service.NewSubscriptionService(subRepo, emailService, weatherClient, appBaseURL)

	// Subscription service schjeduler
	schedulerService := scheduler.NewScheduler(subscriptionSvc)

	weatherUpdateCronSpec := "*/2 * * * *" // every 2 minutes
	if err := schedulerService.SetupAndStartDefaultJobs(weatherUpdateCronSpec); err != nil {
		log.Fatalf("Could not setup and start scheduler jobs: %v", err)
	}

	// API Handlers
	weatherHandler := api.NewWeatherHandler(weatherClient)
	subscriptionHandler := api.NewSubscriptionHandler(subscriptionSvc)

	// Router
	router := api.NewRouter(weatherHandler, subscriptionHandler)

	// Server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", port, err)
	}
}
