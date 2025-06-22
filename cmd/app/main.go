package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go-averroes/internal/common"
	"go-averroes/internal/routes"
	"log"
)

func main() {
	cfg := common.LoadDBConfig()
	if err := common.InitDB(cfg); err != nil {
		log.Fatalf("Erreur de connexion à la base de données : %v", err)
	}

	router := gin.Default()
	routes.RegisterRoutes(router)

	err := router.Run(":8080")
	if err != nil {
		return
	}
}
