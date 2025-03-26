package main

import (
	"custodian/internal/handlers"
	"custodian/internal/pkg"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Specify the environment file and config file locations
	envFile := "config/dev.env" // Change this to the specific .env file you want
	configFile := "config/config.yaml"

	// Load configuration with a specific `.env` file
	config, err := pkg.LoadConfig(configFile, envFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := pkg.NewLogger(config)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	router := gin.Default()

	// Apply API request logging middleware
	router.Use(logger.LogMiddleware())

	// 🔹 Apply CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register routes
	logger.LogMessage("INFO", fmt.Sprintf("🚀 Your data Custodian has started operating in %s mode on port 8080", config.Env))

	// Initialize MongoDB
	mongoDB := pkg.ConnectMongoDB(config.Mongodb.Uri, config.Mongodb.Database)

	// Register routes
	handlers.RegisterRoutes(router)

	// Start server
	router.Run(":8080")

	// Close MongoDB connection on exit
	defer mongoDB.Client.Disconnect(nil)
}
