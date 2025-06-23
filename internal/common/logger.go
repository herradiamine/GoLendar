package common

import (
	"log/slog"
	"os"
)

// InitLogger initialise le logger structuré slog avec sortie fichier et format JSON.
func InitLogger(logLevel slog.Level) error {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err = os.Mkdir("logs", 0755)
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	h := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))
	return nil
}
