package main

import (
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/internal/routes"
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	slog.Info(common.LogAppStart)
	cfg := common.LoadDBConfig()
	if err := common.InitDB(cfg); err != nil {
		log.Fatalf(common.ErrDatabaseConnection, err)
	}

	router := gin.Default()
	router.Use(middleware.LoggingMiddleware())
	routes.RegisterRoutes(router)

	err := router.Run(":8080")
	if err != nil {
		return
	}
}
