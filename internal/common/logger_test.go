package common

import (
	"log/slog"
	"os"
	"testing"
)

func TestInitLogger(t *testing.T) {
	// Nettoyage du fichier de log avant/après
	_ = os.RemoveAll("logs/app.log")
	_ = os.RemoveAll("logs")

	err := InitLogger(slog.LevelInfo)
	if err != nil {
		t.Errorf("InitLogger a retourné une erreur: %v", err)
	}

	// Vérifie que le fichier de log a été créé
	if _, err := os.Stat("logs/app.log"); os.IsNotExist(err) {
		t.Error("Le fichier de log n'a pas été créé")
	}

	// Nettoyage après test
	_ = os.RemoveAll("logs/app.log")
	_ = os.RemoveAll("logs")
}
