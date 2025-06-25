// Package user_calendar internal/user_calendar/user_calendar.go
package user_calendar

import (
	"database/sql"
	"errors"
	"go-averroes/internal/common"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserCalendarStruct struct{}

var UserCalendar = UserCalendarStruct{}

// Get récupère une liaison user calendar par user_id et calendar_id
func (UserCalendarStruct) Get(c *gin.Context) {
	slog.Info(common.LogUserCalendarGet)
	userID, ok := checkUserAccess(c)
	if !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}
	calendarID := calendarData.CalendarID

	var userCalendar common.UserCalendar
	err := common.DB.QueryRow(`
		SELECT user_calendar_id, user_id, calendar_id, created_at, updated_at, deleted_at 
		FROM user_calendar 
		WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL
	`, userID, calendarID).Scan(&userCalendar.UserCalendarID, &userCalendar.UserID, &userCalendar.CalendarID, &userCalendar.CreatedAt, &userCalendar.UpdatedAt, &userCalendar.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error(common.LogUserCalendarGet + " - liaison non trouvée")
			c.JSON(http.StatusNotFound, common.JSONResponse{
				Success: false,
				Error:   common.ErrUserCalendarNotFound,
			})
			return
		}
		slog.Error(common.LogUserCalendarGet + " - erreur SQL : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}

	slog.Info(common.LogUserCalendarGet + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    userCalendar,
	})
}

// Add crée une nouvelle liaison user calendar
func (UserCalendarStruct) Add(c *gin.Context) {
	slog.Info(common.LogUserCalendarAdd)
	userID, ok := checkUserAccess(c)
	if !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}
	calendarID := calendarData.CalendarID

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogUserCalendarAdd + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Vérifier si la liaison existe déjà
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", userID, calendarID).Scan(&existingLink.UserCalendarID)
	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error(common.LogUserCalendarAdd + " - conflit : liaison déjà existante")
		c.JSON(http.StatusConflict, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarConflict,
		})
		return
	}

	// Insérer la liaison
	result, err := tx.Exec(`
		INSERT INTO user_calendar (user_id, calendar_id, created_at) 
		VALUES (?, ?, NOW())
	`, userID, calendarID)
	if err != nil {
		// Vérifier si c'est une erreur de doublon MySQL
		if strings.Contains(err.Error(), "Duplicate entry") {
			slog.Error(common.LogUserCalendarAdd + " - " + common.LogUserCalendarConflict)
			c.JSON(http.StatusConflict, common.JSONResponse{
				Success: false,
				Error:   common.ErrUserCalendarConflict,
			})
			return
		}
		slog.Error(common.LogUserCalendarAdd + " - " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarLinkCreation,
		})
		return
	}

	userCalendarID, _ := result.LastInsertId()

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		slog.Error(common.LogUserCalendarAdd + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogUserCalendarAdd + " - succès")
	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessCreateUserCalendar,
		Data:    gin.H{"user_calendar_id": userCalendarID},
	})
}

// Update met à jour une liaison user calendar par user_id et calendar_id
func (UserCalendarStruct) Update(c *gin.Context) {
	slog.Info(common.LogUserCalendarUpdate)
	userID, ok := checkUserAccess(c)
	if !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}
	calendarID := calendarData.CalendarID

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogUserCalendarUpdate + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	// Vérifier si la liaison existe
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", userID, calendarID).Scan(&existingLink.UserCalendarID)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Error(common.LogUserCalendarUpdate + " - liaison non trouvée")
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarNotFound,
		})
		return
	}

	// Ici, on pourrait mettre à jour d'autres champs si besoin (ex : updated_at).
	_, err = tx.Exec("UPDATE user_calendar SET updated_at = NOW() WHERE user_id = ? AND calendar_id = ?", userID, calendarID)
	if err != nil {
		slog.Error(common.LogUserCalendarUpdate + " - erreur lors de la mise à jour de la liaison : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarUpdate,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error(common.LogUserCalendarUpdate + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogUserCalendarUpdate + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUpdateUserCalendar,
	})
}

// Delete supprime une liaison user-calendar par user_id et calendar_id
func (UserCalendarStruct) Delete(c *gin.Context) {
	slog.Info(common.LogUserCalendarDelete)
	userID, ok := checkUserAccess(c)
	if !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}
	calendarID := calendarData.CalendarID

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogUserCalendarDelete + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	// Vérifier si la liaison existe
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", userID, calendarID).Scan(&existingLink.UserCalendarID)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Error(common.LogUserCalendarDelete + " - liaison non trouvée")
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarNotFound,
		})
		return
	}

	_, err = tx.Exec("UPDATE user_calendar SET deleted_at = NOW() WHERE user_id = ? AND calendar_id = ?", userID, calendarID)
	if err != nil {
		slog.Error(common.LogUserCalendarDelete + " - erreur lors de la suppression de la liaison : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarDelete,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error(common.LogUserCalendarDelete + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogUserCalendarDelete + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessDeleteUserCalendar,
	})
}

// List récupère toutes les liaisons user calendar pour un utilisateur
func (UserCalendarStruct) List(c *gin.Context) {
	slog.Info(common.LogUserCalendarList)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

	rows, err := common.DB.Query(`
		SELECT uc.user_calendar_id, uc.user_id, uc.calendar_id, uc.created_at, uc.updated_at, uc.deleted_at,
		       c.title, c.description
		FROM user_calendar uc
		INNER JOIN calendar c ON uc.calendar_id = c.calendar_id
		WHERE uc.user_id = ? AND uc.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY uc.created_at DESC
	`, userID)
	if err != nil {
		slog.Error(common.LogUserCalendarList + " - erreur lors de la récupération des liaisons : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer rows.Close()

	var userCalendars []common.UserCalendarWithDetails
	for rows.Next() {
		var userCalendar common.UserCalendarWithDetails
		err := rows.Scan(
			&userCalendar.UserCalendarID,
			&userCalendar.UserID,
			&userCalendar.CalendarID,
			&userCalendar.CreatedAt,
			&userCalendar.UpdatedAt,
			&userCalendar.DeletedAt,
			&userCalendar.Title,
			&userCalendar.Description,
		)
		if err != nil {
			slog.Error(common.LogUserCalendarList + " - erreur lors de la lecture des données : " + err.Error())
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   common.ErrTransactionStart,
			})
			return
		}
		userCalendars = append(userCalendars, userCalendar)
	}

	slog.Info(common.LogUserCalendarList + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessListUserCalendars,
		Data:    userCalendars,
	})
}

// GetByUser récupère toutes les liaisons user calendar pour un utilisateur spécifique (admin)
func (UserCalendarStruct) GetByUser(c *gin.Context) {
	slog.Info(common.LogUserCalendarList)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

	rows, err := common.DB.Query(`
		SELECT uc.user_calendar_id, uc.user_id, uc.calendar_id, uc.created_at, uc.updated_at, uc.deleted_at,
		       c.title, c.description
		FROM user_calendar uc
		INNER JOIN calendar c ON uc.calendar_id = c.calendar_id
		WHERE uc.user_id = ? AND uc.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY uc.created_at DESC
	`, userID)
	if err != nil {
		slog.Error(common.LogUserCalendarList + " - erreur lors de la récupération des liaisons : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer rows.Close()

	var userCalendars []common.UserCalendarWithDetails
	for rows.Next() {
		var userCalendar common.UserCalendarWithDetails
		err := rows.Scan(
			&userCalendar.UserCalendarID,
			&userCalendar.UserID,
			&userCalendar.CalendarID,
			&userCalendar.CreatedAt,
			&userCalendar.UpdatedAt,
			&userCalendar.DeletedAt,
			&userCalendar.Title,
			&userCalendar.Description,
		)
		if err != nil {
			slog.Error(common.LogUserCalendarList + " - erreur lors de la lecture des données : " + err.Error())
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   common.ErrTransactionStart,
			})
			return
		}
		userCalendars = append(userCalendars, userCalendar)
	}

	slog.Info(common.LogUserCalendarList + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessListUserCalendars,
		Data:    userCalendars,
	})
}

// checkUserAccess vérifie que l'utilisateur authentifié correspond au user_id de l'URL
func checkUserAccess(c *gin.Context) (int, bool) {
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return 0, false
	}
	authUserID := userData.UserID

	userIDParam := c.Param("user_id")
	if userIDParam == "" {
		slog.Error("[user_calendar][Access] user_id manquant dans l'URL")
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidUserID,
		})
		return 0, false
	}
	userIDFromURL, err := strconv.Atoi(userIDParam)
	if err != nil {
		slog.Error("[user_calendar][Access] user_id invalide dans l'URL")
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidUserID,
		})
		return 0, false
	}
	if authUserID != userIDFromURL {
		slog.Error("[user_calendar][Access] " + common.LogUserCalendarUnauthorizedAccess)
		c.JSON(http.StatusForbidden, common.JSONResponse{
			Success: false,
			Error:   common.ErrNoAccessToCalendar,
		})
		return 0, false
	}
	return userIDFromURL, true
}
