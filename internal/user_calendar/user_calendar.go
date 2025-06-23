// Package user_calendar internal/user_calendar/user_calendar.go
package user_calendar

import (
	"database/sql"
	"errors"
	"go-averroes/internal/common"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserCalendarStruct struct{}

var UserCalendar = UserCalendarStruct{}

// Get récupère une liaison user calendar par user_id et calendar_id
func (UserCalendarStruct) Get(c *gin.Context) {
	slog.Info(common.LogUserCalendarGet)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

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
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

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
		slog.Error(common.LogUserCalendarAdd + " - erreur lors de l'ajout de la liaison : " + err.Error())
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
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

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
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

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
