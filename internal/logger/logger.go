package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

// Init initializes the global logger instance
func Init() error {
	var err error
	log, err = zap.NewProduction() // or zap.NewDevelopment()
	if err != nil {
		return err
	}
	return nil
}

// GetLogger returns the logger instance
func GetLogger() *zap.Logger {
	return log
}

// Sugar returns a sugared logger for convenience
func Sugar() *zap.SugaredLogger {
	return log.Sugar()
}
