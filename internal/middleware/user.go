package middleware

import (
	"database/sql"
	"go-averroes/internal/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserExistsMiddleware vérifie l'existence d'un utilisateur à partir d'un paramètre dans l'URL
// paramName: nom du paramètre à vérifier (ex: "id", "user_id")
func UserExistsMiddleware(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param(paramName)
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   "ID utilisateur invalide",
			})
			c.Abort()
			return
		}

		var user common.User
		err = common.DB.QueryRow(
			"SELECT user_id, lastname, firstname, email, created_at, updated_at, deleted_at FROM user WHERE user_id = ? AND deleted_at IS NULL",
			userID,
		).Scan(
			&user.UserID,
			&user.Lastname,
			&user.Firstname,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, common.JSONResponse{
				Success: false,
				Error:   "Utilisateur non trouvé",
			})
			c.Abort()
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   "Erreur lors de la vérification de l'utilisateur",
			})
			c.Abort()
			return
		}

		// L'utilisateur existe, on l'ajoute au contexte et on continue
		c.Set("user", user)
		c.Next()
	}
}
