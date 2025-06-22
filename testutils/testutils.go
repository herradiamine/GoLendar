package testutils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go-averroes/internal/common"
)

// SetupTestDB configure la base de données de test
func SetupTestDB() {
	config := common.DBConfig{
		User:     "root",
		Password: "password",
		Host:     "localhost",
		Port:     3306,
		Name:     "calendar",
	}

	err := common.InitDB(config)
	if err != nil {
		panic(fmt.Sprintf("Erreur d'initialisation de la base de données: %v", err))
	}
}

// SetupTestRouter configure un router de test basique
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	SetupTestDB()
	return gin.New()
}
