package common

import (
	"database/sql"
	"fmt"
)

var DB *sql.DB

// InitDB initialise la connexion à la base de données MySQL à partir d'une DBConfig
func InitDB(cfg DBConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// Vérifie la connexion
	return DB.Ping()
}
