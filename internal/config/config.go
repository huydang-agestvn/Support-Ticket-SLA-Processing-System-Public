package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Config struct {
	DBHost         string
	DBPort         int
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	ServerPort     int
	WorkerPoolSize int
	DB             *gorm.DB
}

// LoadConfig
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	cfg := &Config{
		// Database: Found environment variables for database configuration
		DBHost:     getEnv("DB_HOST"),
		DBPort:     getEnvInt("DB_PORT"),
		DBUser:     getEnv("DB_USER"),
		DBPassword: getEnv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME"),
		DBSSLMode:  getEnv("DB_SSLMODE"),

		ServerPort:     getEnvInt("SERVER_PORT"),
		WorkerPoolSize: getEnvInt("WORKER_POOL_SIZE"),
	}

	return cfg
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// Giữ nguyên hàm GetDatabase và các helper của bạn...
func (c *Config) GetDatabase() (*gorm.DB, error) {
	if c.DB != nil {
		return c.DB, nil
	}

	var db *gorm.DB
	var err error
	maxRetries := 5
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		// Lưu ý: Bạn cần đảm bảo hàm initDatabase(c) đã được định nghĩa ở file khác trong package config
		db, err = initDatabase(c)
		if err == nil {
			c.DB = db
			return db, nil
		}

		if i < maxRetries-1 {
			fmt.Printf("Failed to connect to database (attempt %d/%d), retrying in %s...\n", i+1, maxRetries, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}
	return nil, fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
}

func getEnv(key string) string {
	value := os.Getenv(key)
	return value
}

func GetPoolSize(key string) int {
	value := os.Getenv(key)
	intVal, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error converting %s to integer: %v. Using default value 5", key, err)
		return 5
	}
	return intVal
}
func GetBatchSize(key string) int {
	value := os.Getenv(key)
	intVal, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error converting %s to integer: %v. Using default value 1000", key, err)
		return 1000
	}
	return intVal
}

func getEnvInt(key string) int {
	value := os.Getenv(key)
	intVal, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Error converting %s to integer: %v", key, err)
	}
	return intVal
}
