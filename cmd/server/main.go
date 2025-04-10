package main

import (
	"custodian/internal/handlers"
	"custodian/internal/logger"
	"custodian/internal/pkg"
	"custodian/internal/service"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Specify the environment file and config file locations
	envFile := "config/dev.env" // Change this to the specific .env file you want
	configFile := "config/config.yaml"

	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.GetLogger().Sync()

	// Load configuration with a specific `.env` file
	config, err := pkg.LoadConfig(configFile, envFile)
	if err != nil {
		logger.GetLogger().Info("Failed to load config file.", zap.String("file", configFile), zap.Error(err))
	}

	router := gin.Default()

	// 🔹 Apply CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001", "https://a8cb2cd1-0b15-4861-810c-d148b964d3a0.e1-us-east-azure.choreoapps.dev",
			"https://7ae7a48f-409c-4152-b389-4e476be31258.e1-us-east-azure.choreoapps.dev", "https://localhost:9001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	logger.GetLogger().Info("Customer Data Service Component has started.", zap.String("env", config.Env))

	// Initialize MongoDB
	mongoDB := pkg.ConnectMongoDB(config.Mongodb.Uri, config.Mongodb.Database)

	pkg.InitLocks(mongoDB.Database)

	// Initialize Event queue
	service.StartEnrichmentWorker()

	// Register routes
	handlers.RegisterRoutes(router)

	// Start server
	router.Run(":8900")

	// Close MongoDB connection on exit
	defer mongoDB.Client.Disconnect(nil)
}
