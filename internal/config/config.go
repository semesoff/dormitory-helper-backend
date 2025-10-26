package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerConfig   ServerConfig
	DatabaseConfig DatabaseConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	JWTSecretKey []byte
}

type DatabaseConfig struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	Driver   string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load() {
	// Загрузка конфигурации сервера
	c.ServerConfig.Host = os.Getenv("SERVER_HOST")
	if c.ServerConfig.Host == "" {
		panic("SERVER_HOST environment variable is required")
	}

	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		panic("SERVER_PORT environment variable is required")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Sprintf("invalid SERVER_PORT: %v", err))
	}
	c.ServerConfig.Port = port

	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		panic("JWT_SECRET_KEY environment variable is required")
	}
	c.ServerConfig.JWTSecretKey = []byte(jwtSecret)

	// Загрузка конфигурации базы данных
	c.DatabaseConfig.Host = os.Getenv("DATABASE_HOST")
	if c.DatabaseConfig.Host == "" {
		panic("DATABASE_HOST environment variable is required")
	}

	dbPortStr := os.Getenv("DATABASE_PORT")
	if dbPortStr == "" {
		panic("DATABASE_PORT environment variable is required")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		panic(fmt.Sprintf("invalid DATABASE_PORT: %v", err))
	}
	c.DatabaseConfig.Port = dbPort

	c.DatabaseConfig.DBName = os.Getenv("DATABASE_NAME")
	if c.DatabaseConfig.DBName == "" {
		panic("DATABASE_NAME environment variable is required")
	}

	c.DatabaseConfig.User = os.Getenv("DATABASE_USER")
	if c.DatabaseConfig.User == "" {
		panic("DATABASE_USER environment variable is required")
	}

	c.DatabaseConfig.Password = os.Getenv("DATABASE_PASSWORD")
	if c.DatabaseConfig.Password == "" {
		panic("DATABASE_PASSWORD environment variable is required")
	}

	c.DatabaseConfig.Driver = os.Getenv("DRIVER")
	if c.DatabaseConfig.Driver == "" {
		panic("DRIVER environment variable is required")
	}
}
