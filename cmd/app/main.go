package main

import (
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/internal/routes"
	"log"
	"log/slog"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if err := common.InitLogger(slog.LevelInfo); err != nil {
		log.Fatalf(common.ErrLoggerInit, err)
	}

	slog.Info(common.LogAppStart)
	cfg := common.LoadDBConfig()
	if err := common.InitDB(cfg); err != nil {
		log.Fatalf(common.ErrDatabaseConnection, err)
	}

	// Configurer Gin en mode debug pour plus de logs
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.Use(location.Default())
	router.Use(middleware.LoggingMiddleware())
	routes.RegisterRoutes(router)

	slog.Info("Serveur démarré sur le port 8080")
	err := router.Run(":8080")
	if err != nil {
		return
	}
}
