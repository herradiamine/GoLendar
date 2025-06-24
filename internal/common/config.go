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

// LoadDBConfig charge la configuration de la base de données depuis les variables d'environnement ou des valeurs par défaut
func LoadDBConfig() DBConfig {
	return DBConfig{
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "password"),
		Host:     getEnv("DB_HOST", "golendar_db"),
		Port:     3306,
		Name:     getEnv("DB_NAME", "calendar"),
	}
}

// getEnv retourne la valeur d'une variable d'environnement ou une valeur par défaut si elle n'est pas définie.
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
