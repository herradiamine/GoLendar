package common

import (
	"os"
)

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
}

// LoadDBConfig charge la configuration depuis les variables d'environnement ou des valeurs par défaut
func LoadDBConfig() DBConfig {
	return DBConfig{
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "password"),
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     3306, // Peut être rendu dynamique si besoin
		Name:     getEnv("DB_NAME", "calendar"),
	}
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
