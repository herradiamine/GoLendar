package common

import (
	"io"
	"log/slog"
	"os"
)

// multiWriter combine plusieurs writers
type multiWriter struct {
	writers []io.Writer
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// InitLogger initialise le logger structuré slog avec sortie fichier et console.
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

	// Créer un writer multi-destination (console + fichier)
	multiWriter := &multiWriter{
		writers: []io.Writer{os.Stdout, file},
	}

	// Créer un handler qui écrit sur les deux destinations
	handler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))
	return nil
}
