package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config structure for reading log settings
type Config struct {
	Env string `yaml:"env"`

	Log struct {
		EnableFileLogging bool   `yaml:"enable_file_logging"`
		LogDirectory      string `yaml:"log_directory"`
		LogFilename       string `yaml:"log_filename"`
		Level             string `yaml:"level"`
	} `yaml:"log"`

	Mongodb struct {
		Uri      string `yaml:"uri"`
		Database string `yaml:"database"`
	} `yaml:"mongodb"`
}

// Logger struct for managing logs
type Logger struct {
	file     *os.File
	logLevel string
}

// LoadConfig reads config from YAML & supports environment variables
// LoadConfig reads YAML and loads a specific `.env` file
func LoadConfig(configPath, envFilePath string) (*Config, error) {
	// Load a specific environment file (instead of default `.env`)
	if envFilePath != "" {
		err := godotenv.Load(envFilePath) // Load custom .env file
		if err != nil {
			log.Printf("Warning: Could not load environment file: %v", err)
		}
	}

	config := &Config{}
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Replace environment variables inside YAML manually
	yamlStr := os.ExpandEnv(string(file))

	err = yaml.Unmarshal([]byte(yamlStr), config)
	if err != nil {
		return nil, err
	}

	// Override YAML values with environment variables if set
	if env := os.Getenv("ENV"); env != "" {
		config.Env = env
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Log.Level = logLevel
	}

	return config, nil
}

// NewLogger initializes a logger with file logging support
func NewLogger(config *Config) (*Logger, error) {
	if !config.Log.EnableFileLogging {
		return &Logger{file: nil, logLevel: strings.ToUpper(config.Log.Level)}, nil
	}

	// Ensure log directory exists
	if err := os.MkdirAll(config.Log.LogDirectory, os.ModePerm); err != nil {
		return nil, err
	}

	// Generate a new filename with timestamp (YYYYMMDD-HHMMSS.log)
	timestamp := time.Now().Format("20060102-150405")
	logFilePath := filepath.Join(config.Log.LogDirectory, fmt.Sprintf("%s-%s", timestamp, config.Log.LogFilename))

	// Open the log file (overwrite if exists)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{file: file, logLevel: strings.ToUpper(config.Log.Level)}, nil
}

// LogMessage logs messages based on log level (DEBUG, INFO, ERROR)
func (l *Logger) LogMessage(level string, message string) {
	allowedLevels := map[string]int{"DEBUG": 1, "INFO": 2, "ERROR": 3}

	// Only log messages that are at or above the configured level
	if allowedLevels[level] < allowedLevels[l.logLevel] {
		return
	}

	logEntry := fmt.Sprintf("%s [%s] [%s] %s", time.Now().Format("2006-01-02 15:04:05"), l.logLevel, level, message)

	// Print to stdout
	fmt.Println(logEntry)

	// Write to log file if enabled
	if l.file != nil {
		l.file.WriteString(logEntry + "\n")
	}
}

// LogMiddleware logs API requests
func (l *Logger) LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		logEntry := fmt.Sprintf(
			"%s [%s] [API] %s %s %d | %s | %s",
			time.Now().Format("2006-01-02 15:04:05"),
			l.logLevel,
			c.Request.Method,
			c.Request.RequestURI,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
		)

		// Print and write log
		fmt.Println(logEntry)
		if l.file != nil {
			l.file.WriteString(logEntry + "\n")
		}
	}
}

// Close closes the log file
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
