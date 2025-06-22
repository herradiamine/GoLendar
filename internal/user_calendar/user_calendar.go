// Package user_calendar internal/user_calendar/user_calendar.go
package user_calendar

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"go-averroes/internal/common"
	"net/http"
	"strconv"
)

type UserCalendarStruct struct{}

var UserCalendar = UserCalendarStruct{}

// Get récupère une liaison user-calendar par son ID
func (UserCalendarStruct) Get(c *gin.Context) {
	id := c.Param("id")
	userCalendarID, err := strconv.Atoi(id)
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
		WHERE user_calendar_id = ? AND deleted_at IS NULL
	`, userCalendarID).Scan(&userCalendar.UserCalendarID, &userCalendar.UserID, &userCalendar.CalendarID, &userCalendar.CreatedAt, &userCalendar.UpdatedAt, &userCalendar.DeletedAt)

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
	var req struct {
		UserID     int `json:"user_id" binding:"required"`
		CalendarID int `json:"calendar_id" binding:"required"`
	}
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

	// Vérifier si le calendrier existe
	var existingCalendar common.Calendar
	err = common.DB.QueryRow("SELECT calendar_id FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL", req.CalendarID).Scan(&existingCalendar.CalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Calendrier non trouvé",
		})
		return
	}

	// Vérifier si la liaison existe déjà
	var existingLink common.UserCalendar
	err = common.DB.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL", req.UserID, req.CalendarID).Scan(&existingLink.UserCalendarID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusConflict, common.JSONResponse{
			Success: false,
			Error:   "Cette liaison utilisateur-calendrier existe déjà",
		})
		return
	}

	// Insérer la liaison
	result, err := common.DB.Exec(`
		INSERT INTO user_calendar (user_id, calendar_id, created_at) 
		VALUES (?, ?, NOW())
	`, req.UserID, req.CalendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la création de la liaison",
		})
		return
	}

	userCalendarID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: "Liaison utilisateur-calendrier créée avec succès",
		Data:    gin.H{"user_calendar_id": userCalendarID},
	})
}

// Update met à jour une liaison user-calendar
func (UserCalendarStruct) Update(c *gin.Context) {
	id := c.Param("id")
	userCalendarID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID liaison invalide",
		})
		return
	}

	var req struct {
		UserID     *int `json:"user_id,omitempty"`
		CalendarID *int `json:"calendar_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

	// Vérifier si la liaison existe
	var existingLink common.UserCalendar
	err = common.DB.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_calendar_id = ? AND deleted_at IS NULL", userCalendarID).Scan(&existingLink.UserCalendarID)
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

	if req.UserID != nil {
		// Vérifier si le nouvel utilisateur existe
		var existingUser common.User
		err = common.DB.QueryRow("SELECT user_id FROM user WHERE user_id = ? AND deleted_at IS NULL", *req.UserID).Scan(&existingUser.UserID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   "Utilisateur non trouvé",
			})
			return
		}
		query += ", user_id = ?"
		args = append(args, *req.UserID)
	}
	if req.CalendarID != nil {
		// Vérifier si le nouveau calendrier existe
		var existingCalendar common.Calendar
		err = common.DB.QueryRow("SELECT calendar_id FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL", *req.CalendarID).Scan(&existingCalendar.CalendarID)
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

	query += " WHERE user_calendar_id = ?"
	args = append(args, userCalendarID)

	_, err = common.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la mise à jour de la liaison",
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
	id := c.Param("id")
	userCalendarID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID liaison invalide",
		})
		return
	}

	// Vérifier si la liaison existe
	var existingLink common.UserCalendar
	err = common.DB.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_calendar_id = ? AND deleted_at IS NULL", userCalendarID).Scan(&existingLink.UserCalendarID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Liaison utilisateur-calendrier non trouvée",
		})
		return
	}

	// Soft delete de la liaison
	_, err = common.DB.Exec("UPDATE user_calendar SET deleted_at = NOW() WHERE user_calendar_id = ?", userCalendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression de la liaison",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: "Liaison utilisateur-calendrier supprimée avec succès",
	})
}
