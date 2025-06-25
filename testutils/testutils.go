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

// ResetTestDB vide toutes les tables de la base de test pour garantir un état propre entre chaque test.
func ResetTestDB() {
	if common.DB == nil {
		// Si la base n'est pas initialisée, on l'initialise d'abord
		SetupTestDB()
		return
	}

	stmts := []string{
		"SET FOREIGN_KEY_CHECKS = 0;",
		"TRUNCATE TABLE user_roles;",
		"TRUNCATE TABLE user_session;",
		"TRUNCATE TABLE user_password;",
		"TRUNCATE TABLE user_calendar;",
		"TRUNCATE TABLE calendar_event;",
		"TRUNCATE TABLE event;",
		"TRUNCATE TABLE calendar;",
		"TRUNCATE TABLE roles;",
		"TRUNCATE TABLE user;",
		"SET FOREIGN_KEY_CHECKS = 1;",
	}

	for _, stmt := range stmts {
		_, err := common.DB.Exec(stmt)
		if err != nil {
			// Si on a une erreur de connexion fermée, on réinitialise la base
			if err.Error() == "sql: database is closed" {
				SetupTestDB()
				// On réessaie une fois après réinitialisation
				_, err = common.DB.Exec(stmt)
				if err != nil {
					panic(fmt.Sprintf("Erreur lors du reset de la base de test après réinitialisation: %v (requête: %s)", err, stmt))
				}
			} else {
				panic(fmt.Sprintf("Erreur lors du reset de la base de test: %v (requête: %s)", err, stmt))
			}
		}
	}
}
