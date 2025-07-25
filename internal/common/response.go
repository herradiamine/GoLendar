package common

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
}

type JSONErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
}

// HandleDBError gère les erreurs courantes de la base de données (ErrNoRows ou autre)
// et envoie la réponse JSON appropriée. Retourne true si une erreur a été gérée.
func HandleDBError(c *gin.Context, err error, statusNotFound int, msgNotFound string, msgInternal string) bool {
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(statusNotFound, JSONResponse{
				Success: false,
				Error:   msgNotFound,
			})
		} else {
			c.JSON(http.StatusInternalServerError, JSONResponse{
				Success: false,
				Error:   msgInternal,
			})
		}
		c.Abort()
		return true
	}
	return false
}
