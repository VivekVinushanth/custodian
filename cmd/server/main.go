package main

import (
	"github.com/joho/godotenv"
	"github.com/wso2/identity-customer-data-service/pkg/handlers"
	"github.com/wso2/identity-customer-data-service/pkg/locks"
	"github.com/wso2/identity-customer-data-service/pkg/logger"
	"github.com/wso2/identity-customer-data-service/pkg/service"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	MongoDB struct {
		URI               string `yaml:"uri"`
		Database          string `yaml:"database"`
		ProfileCollection string `yaml:"profile_collection"`
		EventCollection   string `yaml:"event_collection"`
		ConsentCollection string `yaml:"consent_collection"`
	} `yaml:"mongodb"`
	Addr struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"addr"`
	Log struct {
		debugEnabled bool `yaml:"debug_enabled"`
	} `yaml:"log"`
	Auth struct {
		CORSAllowedOrigins []string `yaml:"cors_allowed_origins"`
	} `yaml:"auth"`
}

func main() {
	const configFile = "config/config.yaml"

	envFiles, err := filepath.Glob("config/*.env")
	if err != nil || len(envFiles) == 0 {
		log.Printf("No .env files found in the config folder: %v", err)
	}
	err = godotenv.Load(envFiles...)

	// Load the configuration
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init(config.Log.debugEnabled)
	router := gin.Default()
	server := handlers.NewServer()

	// 🔹 Apply CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins: config.Auth.CORSAllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		//AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowHeaders:     []string{"*"}, // Or specify "filter" if needed
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	logger.Info("Identity customer data service Component has started.")

	// Initialize MongoDB
	mongoDB := locks.ConnectMongoDB(config.MongoDB.URI, config.MongoDB.Database)

	locks.InitLocks(mongoDB.Database)

	// Initialize Event queue
	service.StartEnrichmentWorker()

	basePath := "/api/v1"
	api := router.Group(basePath)
	handlers.RegisterHandlers(api, server)
	s := &http.Server{
		Handler: router,
		Addr:    config.Addr.Host + ":" + config.Addr.Port,
	}

	// Close MongoDB connection on exit
	defer mongoDB.Client.Disconnect(nil)

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Replace placeholders with environment variables
	expanded := os.ExpandEnv(string(file))

	var config Config
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, err
	}
	return &config, nil
}
