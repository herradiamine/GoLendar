package middleware

import (
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
				Error:   common.ErrInvalidUserID,
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

		if common.HandleDBError(c, err, http.StatusNotFound, common.ErrUserNotFound, common.ErrUserVerification) {
			return
		}

		// L'utilisateur existe, on l'ajoute au contexte et on continue
		c.Set("user", user)
		c.Next()
	}
}

func CalendarExistsMiddleware(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		calendarIDStr := c.Param(paramName)
		calendarID, err := strconv.Atoi(calendarIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   "ID calendrier invalide",
			})
			c.Abort()
			return
		}

		var calendar common.Calendar
		err = common.DB.QueryRow(
			"SELECT calendar_id, title, description, created_at, updated_at, deleted_at FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL",
			calendarID,
		).Scan(
			&calendar.CalendarID,
			&calendar.Title,
			&calendar.Description,
			&calendar.CreatedAt,
			&calendar.UpdatedAt,
			&calendar.DeletedAt,
		)

		if common.HandleDBError(c, err, http.StatusNotFound, common.ErrCalendarNotFound, common.ErrCalendarVerification) {
			return
		}

		// Le calendrier existe, on l'ajoute au contexte et on continue
		c.Set("calendar", calendar)
		c.Next()
	}
}

func UserCanAccessCalendarMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userData, ok := common.GetUserFromContext(c)
		if !ok {
			return
		}

		calendarData, ok := common.GetCalendarFromContext(c)
		if !ok {
			return
		}

		// Vérifier que l'utilisateur a accès au calendrier
		var accessCheck int
		err := common.DB.QueryRow(`
			SELECT 1 FROM user_calendar 
			WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL
		`, userData.UserID, calendarData.CalendarID).Scan(&accessCheck)

		if common.HandleDBError(c, err, http.StatusForbidden, common.ErrNoAccessToCalendar, common.ErrCalendarAccessCheck) {
			return
		}

		c.Next()
	}
}
