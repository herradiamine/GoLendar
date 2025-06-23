// Package user_calendar internal/user_calendar/user_calendar.go
package user_calendar

import (
	"database/sql"
	"go-averroes/internal/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserCalendarStruct struct{}

var UserCalendar = UserCalendarStruct{}

// Get récupère une liaison user-calendar par user_id et calendar_id
func (UserCalendarStruct) Get(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: utilisateur non trouvé dans le contexte",
		})
		return
	}
	userData := user.(common.User)
	userID := userData.UserID

	calendar, exists := c.Get("calendar")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: calendrier non trouvé dans le contexte",
		})
		return
	}
	calendarData := calendar.(common.Calendar)
	calendarID := calendarData.CalendarID

	var userCalendar common.UserCalendar
	err := common.DB.QueryRow(`
		SELECT user_calendar_id, user_id, calendar_id, created_at, updated_at, deleted_at 
		FROM user_calendar 
		WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL
	`, userID, calendarID).Scan(&userCalendar.UserCalendarID, &userCalendar.UserID, &userCalendar.CalendarID, &userCalendar.CreatedAt, &userCalendar.UpdatedAt, &userCalendar.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, common.JSONResponse{
				Success: false,
				Error:   "Liaison utilisateur-calendrier non trouvée",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la récupération de la liaison",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    userCalendar,
	})
}

// Add crée une nouvelle liaison user-calendar
func (UserCalendarStruct) Add(c *gin.Context) {
	// Récupérer l'utilisateur du contexte (déjà vérifié par le middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: utilisateur non trouvé dans le contexte",
		})
		return
	}
	userData := user.(common.User)
	userID := userData.UserID

	// Récupérer le calendrier du contexte (déjà vérifié par le middleware)
	calendar, exists := c.Get("calendar")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: calendrier non trouvé dans le contexte",
		})
		return
	}
	calendarData := calendar.(common.Calendar)
	calendarID := calendarData.CalendarID

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors du démarrage de la transaction",
		})
		return
	}
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Vérifier si la liaison existe déjà
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", userID, calendarID).Scan(&existingLink.UserCalendarID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusConflict, common.JSONResponse{
			Success: false,
			Error:   "Cette liaison utilisateur-calendrier existe déjà",
		})
		return
	}

	// Insérer la liaison
	result, err := tx.Exec(`
		INSERT INTO user_calendar (user_id, calendar_id, created_at) 
		VALUES (?, ?, NOW())
	`, userID, calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la création de la liaison",
		})
		return
	}

	userCalendarID, _ := result.LastInsertId()

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la validation de la transaction",
		})
		return
	}

	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: "Liaison utilisateur-calendrier créée avec succès",
		Data:    gin.H{"user_calendar_id": userCalendarID},
	})
}

// Update met à jour une liaison user-calendar par user_id et calendar_id
func (UserCalendarStruct) Update(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: utilisateur non trouvé dans le contexte",
		})
		return
	}
	userData := user.(common.User)
	userID := userData.UserID

	calendar, exists := c.Get("calendar")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: calendrier non trouvé dans le contexte",
		})
		return
	}
	calendarData := calendar.(common.Calendar)
	calendarID := calendarData.CalendarID

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors du démarrage de la transaction",
		})
		return
	}
	defer tx.Rollback()

	// Vérifier si la liaison existe
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", userID, calendarID).Scan(&existingLink.UserCalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Liaison utilisateur-calendrier non trouvée",
		})
		return
	}

	// Ici, on pourrait mettre à jour d'autres champs si besoin (ex: updated_at)
	_, err = tx.Exec("UPDATE user_calendar SET updated_at = NOW() WHERE user_id = ? AND calendar_id = ?", userID, calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la mise à jour de la liaison",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la validation de la transaction",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: "Liaison utilisateur-calendrier mise à jour avec succès",
	})
}

// Delete supprime une liaison user-calendar par user_id et calendar_id
func (UserCalendarStruct) Delete(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: utilisateur non trouvé dans le contexte",
		})
		return
	}
	userData := user.(common.User)
	userID := userData.UserID

	calendar, exists := c.Get("calendar")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: calendrier non trouvé dans le contexte",
		})
		return
	}
	calendarData := calendar.(common.Calendar)
	calendarID := calendarData.CalendarID

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors du démarrage de la transaction",
		})
		return
	}
	defer tx.Rollback()

	// Vérifier si la liaison existe
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", userID, calendarID).Scan(&existingLink.UserCalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Liaison utilisateur-calendrier non trouvée",
		})
		return
	}

	_, err = tx.Exec("UPDATE user_calendar SET deleted_at = NOW() WHERE user_id = ? AND calendar_id = ?", userID, calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression de la liaison",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la validation de la transaction",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: "Liaison utilisateur-calendrier supprimée avec succès",
	})
}
