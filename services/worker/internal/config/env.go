package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Load attempts to load env paths
func Load() {
	paths := []string{".env", "../../.env", "../../../.env", "../../../../.env"}
	for _, p := range paths {
		_ = godotenv.Load(p)
	}
}

// GetEnv retrieves the value or fatals
func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return value
}

// GetDBConnStr constructs psycopg conn
func GetDBConnStr() string {
	user := GetEnv("POSTGRES_USER")
	password := GetEnv("POSTGRES_PASSWORD")
	dbname := GetEnv("POSTGRES_DB")
	host := GetEnv("POSTGRES_HOST")
	port := GetEnv("POSTGRES_PORT")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}

// GetRabbitMQConnStr constructs amqps conn
func GetRabbitMQConnStr() string {
	user := GetEnv("RABBITMQ_USER")
	password := GetEnv("RABBITMQ_PASSWORD")
	host := GetEnv("RABBITMQ_HOST")
	port := GetEnv("RABBITMQ_PORT")

	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
}
