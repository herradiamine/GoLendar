package testutils

import (
	"fmt"
	"go-averroes/internal/common"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// SetupTestDB configure la base de données de test pour les tests unitaires et d'intégration.
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
		panic(fmt.Sprintf(common.ErrTestDBInit, err))
	}
}

// SetupTestRouter configure un router Gin pour les tests.
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	SetupTestDB()
	return gin.New()
}
