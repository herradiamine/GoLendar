// Package user_calendar internal/user_calendar/user_calendar.go
package user_calendar

import (
	"database/sql"
	"go-averroes/internal/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserCalendarStruct struct{}

var UserCalendar = UserCalendarStruct{}

// Get récupère une liaison user-calendar par son ID
func (UserCalendarStruct) Get(c *gin.Context) {
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

	// Récupérer l'ID de la liaison user_calendar
	userCalendarIDStr := c.Param("user_calendar_id")
	userCalendarID, err := strconv.Atoi(userCalendarIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID liaison invalide",
		})
		return
	}

	var userCalendar common.UserCalendar
	err = common.DB.QueryRow(`
		SELECT user_calendar_id, user_id, calendar_id, created_at, updated_at, deleted_at 
		FROM user_calendar 
		WHERE user_calendar_id = ? AND user_id = ? AND deleted_at IS NULL
	`, userCalendarID, userID).Scan(&userCalendar.UserCalendarID, &userCalendar.UserID, &userCalendar.CalendarID, &userCalendar.CreatedAt, &userCalendar.UpdatedAt, &userCalendar.DeletedAt)

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

	var req struct {
		CalendarID int `json:"calendar_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

	if req.CalendarID <= 0 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "calendar_id doit être positif",
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
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Vérifier si le calendrier existe
	var existingCalendar common.Calendar
	err = tx.QueryRow("SELECT calendar_id FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL", req.CalendarID).Scan(&existingCalendar.CalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Calendrier non trouvé",
		})
		return
	}

	// Vérifier si la liaison existe déjà
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", userID, req.CalendarID).Scan(&existingLink.UserCalendarID)
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
	`, userID, req.CalendarID)
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

// Update met à jour une liaison user-calendar
func (UserCalendarStruct) Update(c *gin.Context) {
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

	// Récupérer l'ID de la liaison user_calendar
	userCalendarIDStr := c.Param("user_calendar_id")
	userCalendarID, err := strconv.Atoi(userCalendarIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID liaison invalide",
		})
		return
	}

	var req struct {
		CalendarID *int `json:"calendar_id,omitempty"`
	}
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
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Vérifier si la liaison existe et appartient à l'utilisateur
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_calendar_id = ? AND user_id = ? AND deleted_at IS NULL", userCalendarID, userID).Scan(&existingLink.UserCalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Liaison utilisateur-calendrier non trouvée",
		})
		return
	}

	// Construire la requête de mise à jour
	query := "UPDATE user_calendar SET updated_at = NOW()"
	var args []interface{}

	if req.CalendarID != nil {
		// Vérifier si le nouveau calendrier existe
		var existingCalendar common.Calendar
		err = tx.QueryRow("SELECT calendar_id FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL", *req.CalendarID).Scan(&existingCalendar.CalendarID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   "Calendrier non trouvé",
			})
			return
		}
		query += ", calendar_id = ?"
		args = append(args, *req.CalendarID)
	}

	query += " WHERE user_calendar_id = ? AND user_id = ?"
	args = append(args, userCalendarID, userID)

	_, err = tx.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la mise à jour de la liaison",
		})
		return
	}

	// Valider la transaction
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

// Delete supprime une liaison user-calendar (soft delete)
func (UserCalendarStruct) Delete(c *gin.Context) {
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

	// Récupérer l'ID de la liaison user_calendar
	userCalendarIDStr := c.Param("user_calendar_id")
	userCalendarID, err := strconv.Atoi(userCalendarIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID liaison invalide",
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
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Vérifier si la liaison existe et appartient à l'utilisateur
	var existingLink common.UserCalendar
	err = tx.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_calendar_id = ? AND user_id = ? AND deleted_at IS NULL", userCalendarID, userID).Scan(&existingLink.UserCalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Liaison utilisateur-calendrier non trouvée",
		})
		return
	}

	// Soft delete de la liaison
	_, err = tx.Exec("UPDATE user_calendar SET deleted_at = NOW() WHERE user_calendar_id = ? AND user_id = ?", userCalendarID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression de la liaison",
		})
		return
	}

	// Valider la transaction
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
