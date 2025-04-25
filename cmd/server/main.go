package main

import (
	"github.com/wso2/identity-customer-data-service/docs"
	"github.com/wso2/identity-customer-data-service/pkg/handlers"
	"github.com/wso2/identity-customer-data-service/pkg/locks"
	"github.com/wso2/identity-customer-data-service/pkg/logger"
	"github.com/wso2/identity-customer-data-service/pkg/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Specify the environment file and config file locations
	//envFile := "config/dev.env" // Change this to the specific .env file you want
	//configFile := "config/config.yaml"

	// Initialize logger

	logger.Init()
	router := gin.Default()
	server := handlers.NewServer()

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

	logger.Log.Info("Identity customer data service Component has started.")

	// Initialize MongoDB
	mongoDB := locks.ConnectMongoDB("mongodb+srv://sa:Q8n%23FUpTpTkpK4%25@cdscluster1.b3chj.mongodb.net/?retryWrites=true&w=majority&appName=cdsCluster1", "custodian_db")

	locks.InitLocks(mongoDB.Database)

	// Initialize Event queue
	service.StartEnrichmentWorker()

	basePath := "/api/v1"
	api := router.Group(basePath)
	docs.RegisterHandlers(api, server)
	s := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8900",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())

	// Close MongoDB connection on exit
	defer mongoDB.Client.Disconnect(nil)
}
