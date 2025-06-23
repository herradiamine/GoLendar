// Package calendar internal/calendar/calendar.go
package calendar

import (
	"database/sql"
	"go-averroes/internal/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CalendarStruct struct{}

var Calendar = CalendarStruct{}

// Get récupère un calendrier par son ID
func (CalendarStruct) Get(c *gin.Context) {
	id := c.Param("id")
	calendarID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID calendrier invalide",
		})
		return
	}

	var calendar common.Calendar
	err = common.DB.QueryRow(`
        SELECT calendar_id, title, description, created_at, updated_at, deleted_at 
        FROM calendar 
        WHERE calendar_id = ? AND deleted_at IS NULL
    `, calendarID).Scan(&calendar.CalendarID, &calendar.Title, &calendar.Description, &calendar.CreatedAt, &calendar.UpdatedAt, &calendar.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, common.JSONResponse{
				Success: false,
				Error:   "Calendrier non trouvé",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la récupération du calendrier",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    calendar,
	})
}

// Add crée un nouveau calendrier
func (CalendarStruct) Add(c *gin.Context) {
	var req common.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

	// Vérifier si l'utilisateur existe
	var existingUser common.User
	err := common.DB.QueryRow("SELECT user_id FROM user WHERE user_id = ? AND deleted_at IS NULL", req.UserID).Scan(&existingUser.UserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Utilisateur non trouvé",
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
    `, req.UserID, calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la création de la liaison utilisateur-calendrier",
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

	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: "Calendrier créé avec succès",
		Data: gin.H{
			"calendar_id": calendarID,
			"user_id":     req.UserID,
		},
	})
}

// Update met à jour un calendrier
func (CalendarStruct) Update(c *gin.Context) {
	id := c.Param("id")
	calendarID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID calendrier invalide",
		})
		return
	}

	var req common.UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

	// Vérifier si le calendrier existe
	var existingCalendar common.Calendar
	err = common.DB.QueryRow("SELECT calendar_id FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL", calendarID).Scan(&existingCalendar.CalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Calendrier non trouvé",
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

	_, err = common.DB.Exec(query, args...)
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

// Delete supprime un calendrier (soft delete)
func (CalendarStruct) Delete(c *gin.Context) {
	id := c.Param("id")
	calendarID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID calendrier invalide",
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
	err = tx.QueryRow("SELECT calendar_id FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL", calendarID).Scan(&existingCalendar.CalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Calendrier non trouvé",
		})
		return
	}

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
		Message: "Calendrier supprimé avec succès",
	})
}
