// backend/internal/config/test_config.go
package config

import (
	"fmt"
	"os"
)

// TestDBConfig representa a configuração do banco de dados para testes
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// LoadTestDBConfig carrega a configuração do banco de dados de teste
func LoadTestDBConfig() TestDBConfig {
	return TestDBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		User:     getEnvOrDefault("TEST_DB_USER", "erp_user"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "changeme"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "erp_test"),
	}
}

// DSN retorna a string de conexão para o banco de dados de teste
func (c TestDBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
