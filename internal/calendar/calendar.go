// Package calendar internal/calendar/calendar.go
package calendar

import (
	"go-averroes/internal/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CalendarStruct struct{}

var Calendar = CalendarStruct{}

// Get récupère un calendrier par user_id et calendar_id
func (CalendarStruct) Get(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: utilisateur non trouvé dans le contexte",
		})
		return
	}
	_ = user.(common.User) // Pour l'instant, on ne l'utilise pas mais il est vérifié

	calendar, exists := c.Get("calendar")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: calendrier non trouvé dans le contexte",
		})
		return
	}
	calendarData := calendar.(common.Calendar)

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    calendarData,
	})
}

// Add crée un nouveau calendrier pour un utilisateur
func (CalendarStruct) Add(c *gin.Context) {
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

	var req common.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

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

	// Insérer le calendrier
	result, err := tx.Exec(`
        INSERT INTO calendar (title, description, created_at) 
        VALUES (?, ?, NOW())
    `, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la création du calendrier",
		})
		return
	}

	calendarID, _ := result.LastInsertId()

	// Créer la liaison user_calendar (propriétaire)
	_, err = tx.Exec(`
        INSERT INTO user_calendar (user_id, calendar_id, created_at) 
        VALUES (?, ?, NOW())
    `, userID, calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la création de la liaison utilisateur-calendrier",
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

	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: "Calendrier créé avec succès",
		Data: gin.H{
			"calendar_id": calendarID,
			"user_id":     userID,
		},
	})
}

// Update met à jour un calendrier par user_id et calendar_id
func (CalendarStruct) Update(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: utilisateur non trouvé dans le contexte",
		})
		return
	}
	_ = user.(common.User)

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

	var req common.UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

	// Construire la requête de mise à jour
	query := "UPDATE calendar SET updated_at = NOW()"
	var args []interface{}

	if req.Title != nil {
		query += ", title = ?"
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		query += ", description = ?"
		args = append(args, *req.Description)
	}

	query += " WHERE calendar_id = ?"
	args = append(args, calendarID)

	_, err := common.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la mise à jour du calendrier",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: "Calendrier mis à jour avec succès",
	})
}

// Delete supprime un calendrier par user_id et calendar_id
func (CalendarStruct) Delete(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur interne: utilisateur non trouvé dans le contexte",
		})
		return
	}
	_ = user.(common.User)

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

	// Soft delete du calendrier
	_, err = tx.Exec("UPDATE calendar SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression du calendrier",
		})
		return
	}

	// Soft delete des liaisons user_calendar
	_, err = tx.Exec("UPDATE user_calendar SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression des liaisons utilisateur-calendrier",
		})
		return
	}

	// Soft delete des liaisons calendar_event
	_, err = tx.Exec("UPDATE calendar_event SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression des liaisons calendrier-événement",
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
		Message: "Calendrier supprimé avec succès",
	})
}
