package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
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
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("failed to load .env file: %v", err))
	}

	// Загрузка конфигурации сервера
	c.ServerConfig.Host = os.Getenv("SERVER_HOST")
	if c.ServerConfig.Host == "" {
		panic("SERVER_HOST is required in .env file")
	}

	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		panic("SERVER_PORT is required in .env file")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Sprintf("invalid SERVER_PORT: %v", err))
	}
	c.ServerConfig.Port = port

	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		panic("JWT_SECRET_KEY is required in .env file")
	}
	c.ServerConfig.JWTSecretKey = []byte(jwtSecret)

	// Загрузка конфигурации базы данных
	c.DatabaseConfig.Host = os.Getenv("DATABASE_HOST")
	if c.DatabaseConfig.Host == "" {
		panic("DATABASE_HOST is required in .env file")
	}

	dbPortStr := os.Getenv("DATABASE_PORT")
	if dbPortStr == "" {
		panic("DATABASE_PORT is required in .env file")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		panic(fmt.Sprintf("invalid DATABASE_PORT: %v", err))
	}
	c.DatabaseConfig.Port = dbPort

	c.DatabaseConfig.DBName = os.Getenv("DATABASE_NAME")
	if c.DatabaseConfig.DBName == "" {
		panic("DATABASE_NAME is required in .env file")
	}

	c.DatabaseConfig.User = os.Getenv("DATABASE_USER")
	if c.DatabaseConfig.User == "" {
		panic("DATABASE_USER is required in .env file")
	}

	c.DatabaseConfig.Password = os.Getenv("DATABASE_PASSWORD")
	if c.DatabaseConfig.Password == "" {
		panic("DATABASE_PASSWORD is required in .env file")
	}

	c.DatabaseConfig.Driver = os.Getenv("DRIVER")
	if c.DatabaseConfig.Driver == "" {
		panic("DRIVER is required in .env file")
	}
}
