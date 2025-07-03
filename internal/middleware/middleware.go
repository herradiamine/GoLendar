package middleware

import (
	"fmt"
	"go-averroes/internal/common"
	"go-averroes/internal/session"
	"net/http"
	"strconv"
	"strings"

	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configure les en-têtes CORS pour permettre la communication avec le frontend
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Autoriser l'origine du frontend
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")

		// Autoriser les méthodes HTTP
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")

		// Autoriser les en-têtes
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Autoriser les credentials (cookies, tokens, etc.)
		c.Header("Access-Control-Allow-Credentials", "true")

		// Gérer les requêtes preflight OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

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
			slog.Error("CalendarExistsMiddleware: ID de calendrier invalide", "calendar_id", calendarIDStr, "error", err.Error())
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   common.ErrInvalidCalendarID,
			})
			c.Abort()
			return
		}

		slog.Info("CalendarExistsMiddleware: Recherche du calendrier", "calendar_id", calendarID)

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

		if err != nil {
			slog.Error("CalendarExistsMiddleware: Calendrier non trouvé", "calendar_id", calendarID, "error", err.Error())
			if common.HandleDBError(c, err, http.StatusNotFound, common.ErrCalendarNotFound, common.ErrCalendarVerification) {
				return
			}
		}

		slog.Info("CalendarExistsMiddleware: Calendrier trouvé", "calendar_id", calendar.CalendarID, "title", calendar.Title)

		// Le calendrier existe, on l'ajoute au contexte et on continue
		c.Set("calendar", calendar)
		c.Next()
	}
}

func RoleExistsMiddleware(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDStr := c.Param(paramName)
		roleID, err := strconv.Atoi(roleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   common.ErrInvalidData,
			})
			c.Abort()
			return
		}

		var role common.Role
		err = common.DB.QueryRow(
			"SELECT role_id, name, description, created_at, updated_at, deleted_at FROM roles WHERE role_id = ? AND deleted_at IS NULL",
			roleID,
		).Scan(
			&role.RoleID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
			&role.DeletedAt,
		)

		if common.HandleDBError(c, err, http.StatusNotFound, common.ErrRoleNotFound, common.ErrRoleNotFound) {
			return
		}

		// Le rôle existe, on l'ajoute au contexte et on continue
		c.Set("role", role)
		c.Next()
	}
}

func UserCanAccessCalendarMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userData, ok := common.GetUserFromContext(c)
		if !ok {
			slog.Error("UserCanAccessCalendarMiddleware: Utilisateur non trouvé dans le contexte")
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrUserNotAuthenticated,
			})
			c.Abort()
			return
		}

		slog.Info("UserCanAccessCalendarMiddleware: Utilisateur trouvé", "user_id", userData.UserID, "email", userData.Email)

		calendarData, ok := common.GetCalendarFromContext(c)
		if !ok {
			slog.Error("UserCanAccessCalendarMiddleware: Calendrier non trouvé dans le contexte")
			c.JSON(http.StatusNotFound, common.JSONResponse{
				Success: false,
				Error:   common.ErrCalendarNotFound,
			})
			c.Abort()
			return
		}

		slog.Info("UserCanAccessCalendarMiddleware: Calendrier trouvé", "calendar_id", calendarData.CalendarID, "title", calendarData.Title)

		// Vérifier que l'utilisateur a accès au calendrier
		var accessCheck int
		err := common.DB.QueryRow(`
			SELECT 1 FROM user_calendar 
			WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL
		`, userData.UserID, calendarData.CalendarID).Scan(&accessCheck)

		if err != nil {
			slog.Error("UserCanAccessCalendarMiddleware: Accès refusé", "user_id", userData.UserID, "calendar_id", calendarData.CalendarID, "error", err.Error())
			// Si le calendrier existe mais pas d'accès, retourner explicitement 403
			c.JSON(http.StatusForbidden, common.JSONResponse{
				Success: false,
				Error:   common.ErrNoAccessToCalendar,
			})
			c.Abort()
			return
		}

		slog.Info("UserCanAccessCalendarMiddleware: Accès autorisé", "user_id", userData.UserID, "calendar_id", calendarData.CalendarID)
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		slog.Info(common.LogHTTPReceivedRequest,
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.String("ip", clientIP),
			slog.Duration("latency", latency),
		)
	}
}

// AuthMiddleware vérifie l'authentification de l'utilisateur via le token de session
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer le token depuis le header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Error(common.LogMissingAuthHeader)
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrUserNotAuthenticated,
			})
			c.Abort()
			return
		}

		// Extraire le token (format: "Bearer <token>")
		token := extractTokenFromHeader(authHeader)
		if token == "" {
			slog.Error(common.LogInvalidToken)
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrSessionInvalid,
			})
			c.Abort()
			return
		}

		// Valider la session
		user, err := session.Session.ValidateSession(token)
		if err != nil {
			slog.Error(common.LogInvalidSession + ": " + err.Error())
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrSessionInvalid,
			})
			c.Abort()
			return
		}

		// Ajouter l'utilisateur au contexte
		c.Set("auth_user", *user)
		c.Next()
	}
}

// RoleMiddleware vérifie que l'utilisateur a un rôle spécifique
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userData, ok := common.GetUserFromContext(c)
		if !ok {
			slog.Error(common.LogUserNotFoundInContext)
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrUserNotAuthenticated,
			})
			c.Abort()
			return
		}

		// Vérifier si l'utilisateur a le rôle requis
		var roleID int
		err := common.DB.QueryRow(`
			SELECT r.role_id
			FROM roles r
			INNER JOIN user_roles ur ON r.role_id = ur.role_id
			WHERE ur.user_id = ? AND r.name = ? AND ur.deleted_at IS NULL AND r.deleted_at IS NULL
		`, userData.UserID, requiredRole).Scan(&roleID)

		if err != nil {
			slog.Error(fmt.Sprintf(common.LogUserMissingRole, requiredRole))
			c.JSON(http.StatusForbidden, common.JSONResponse{
				Success: false,
				Error:   common.ErrInsufficientPermissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RolesMiddleware vérifie que l'utilisateur a au moins un des rôles spécifiés
func RolesMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userData, ok := common.GetUserFromContext(c)
		if !ok {
			slog.Error(common.LogUserNotFoundInContext)
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrUserNotAuthenticated,
			})
			c.Abort()
			return
		}

		// Construire la requête pour vérifier si l'utilisateur a au moins un des rôles
		placeholders := make([]string, len(requiredRoles))
		args := make([]interface{}, len(requiredRoles)+1)
		args[0] = userData.UserID

		for i, role := range requiredRoles {
			placeholders[i] = "?"
			args[i+1] = role
		}

		query := `
			SELECT COUNT(*) 
			FROM roles r
			INNER JOIN user_roles ur ON r.role_id = ur.role_id
			WHERE ur.user_id = ? AND r.name IN (` + strings.Join(placeholders, ",") + `) 
			AND ur.deleted_at IS NULL AND r.deleted_at IS NULL
		`

		var count int
		err := common.DB.QueryRow(query, args...).Scan(&count)

		if err != nil || count == 0 {
			slog.Error("Utilisateur n'a aucun des rôles requis: " + strings.Join(requiredRoles, ", "))
			c.JSON(http.StatusForbidden, common.JSONResponse{
				Success: false,
				Error:   common.ErrInsufficientPermissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware vérifie que l'utilisateur a le rôle "admin"
func AdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// OptionalAuthMiddleware vérifie l'authentification si un token est fourni, sinon continue sans authentification
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer le token depuis le header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Pas de token, on continue sans authentification
			c.Next()
			return
		}

		// Extraire le token
		token := extractTokenFromHeader(authHeader)
		if token == "" {
			// Token invalide, on continue sans authentification
			c.Next()
			return
		}

		// Valider la session
		user, err := session.Session.ValidateSession(token)
		if err != nil {
			// Session invalide, on continue sans authentification
			slog.Warn(common.LogSessionInvalidOptional + ": " + err.Error())
			c.Next()
			return
		}

		// Ajouter l'utilisateur au contexte
		c.Set("auth_user", *user)
		c.Next()
	}
}

// extractTokenFromHeader extrait le token du header Authorization
func extractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// EventExistsMiddleware vérifie l'existence d'un événement à partir d'un paramètre dans l'URL
// paramName: nom du paramètre à vérifier (ex: "id", "event_id")
func EventExistsMiddleware(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventIDStr := c.Param(paramName)
		eventID, err := strconv.Atoi(eventIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   common.ErrInvalidEventID,
			})
			c.Abort()
			return
		}

		var event common.Event
		err = common.DB.QueryRow(
			"SELECT event_id, title, description, start, duration, canceled, created_at, updated_at, deleted_at FROM event WHERE event_id = ? AND deleted_at IS NULL",
			eventID,
		).Scan(
			&event.EventID,
			&event.Title,
			&event.Description,
			&event.Start,
			&event.Duration,
			&event.Canceled,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.DeletedAt,
		)

		if common.HandleDBError(c, err, http.StatusNotFound, common.ErrEventNotFound, common.ErrEventRetrieval) {
			return
		}

		// L'événement existe, on l'ajoute au contexte et on continue
		c.Set("event", event)
		c.Next()
	}
}
