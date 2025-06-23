// Package calendar_event internal/calendar_event/calendar_event.go
package calendar_event

import (
	"database/sql"
	"go-averroes/internal/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CalendarEventStruct struct{}

var CalendarEvent = CalendarEventStruct{}

// Get récupère un événement par son ID
func (CalendarEventStruct) Get(c *gin.Context) {
	_ = c.MustGet("user").(common.User)
	_ = c.MustGet("calendar").(common.Calendar)
	id := c.Param("event_id")
	eventID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidEventID,
		})
		return
	}

	var event common.Event
	err = common.DB.QueryRow(`
		SELECT event_id, title, description, start, duration, canceled, created_at, updated_at, deleted_at 
		FROM event 
		WHERE event_id = ? AND deleted_at IS NULL
	`, eventID).Scan(&event.EventID, &event.Title, &event.Description, &event.Start, &event.Duration, &event.Canceled, &event.CreatedAt, &event.UpdatedAt, &event.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, common.JSONResponse{
				Success: false,
				Error:   common.ErrEventNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la récupération de l'événement",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    event,
	})
}

// Add crée un nouvel événement
func (CalendarEventStruct) Add(c *gin.Context) {
	user := c.MustGet("user").(common.User)
	userID := user.UserID
	calendar := c.MustGet("calendar").(common.Calendar)
	calendarID := calendar.CalendarID

	var req common.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	// Validation des données
	if req.Duration < 1 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidDuration,
		})
		return
	}

	// Vérifier que l'utilisateur a accès au calendrier (propriétaire ou lié via user_calendar)
	var accessCheck int
	err := common.DB.QueryRow(`
		SELECT 1 FROM (
			SELECT user_id FROM calendar WHERE calendar_id = ? AND deleted_at IS NULL
			UNION
			SELECT user_id FROM user_calendar WHERE calendar_id = ? AND deleted_at IS NULL
		) AS access WHERE user_id = ?
	`, calendarID, calendarID, userID).Scan(&accessCheck)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusForbidden, common.JSONResponse{
			Success: false,
			Error:   "Vous n'avez pas accès à ce calendrier",
		})
		return
	}

	// Valeur par défaut pour canceled si non fournie
	canceled := false
	if req.Canceled != nil {
		canceled = *req.Canceled
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

	// Insérer l'événement
	result, err := tx.Exec(`
		INSERT INTO event (title, description, start, duration, canceled, created_at) 
		VALUES (?, ?, ?, ?, ?, NOW())
	`, req.Title, req.Description, req.Start, req.Duration, canceled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventCreation,
		})
		return
	}

	eventID, _ := result.LastInsertId()

	// Créer la liaison calendar_event
	_, err = tx.Exec(`
		INSERT INTO calendar_event (calendar_id, event_id, created_at) 
		VALUES (?, ?, NOW())
	`, calendarID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarEventLink,
		})
		return
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessCreateEvent,
		Data: gin.H{
			"event_id":    eventID,
			"calendar_id": calendarID,
		},
	})
}

// Update met à jour un événement
func (CalendarEventStruct) Update(c *gin.Context) {
	_ = c.MustGet("user").(common.User)
	_ = c.MustGet("calendar").(common.Calendar)
	id := c.Param("event_id")
	eventID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidEventID,
		})
		return
	}

	var req common.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	// Validation des données
	if req.Duration != nil && *req.Duration < 1 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidDuration,
		})
		return
	}

	// Vérifier si l'événement existe
	var existingEvent common.Event
	err = common.DB.QueryRow("SELECT event_id FROM event WHERE event_id = ? AND deleted_at IS NULL", eventID).Scan(&existingEvent.EventID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventNotFound,
		})
		return
	}

	// Construire la requête de mise à jour
	query := "UPDATE event SET updated_at = NOW()"
	var args []interface{}

	if req.Title != nil {
		query += ", title = ?"
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		query += ", description = ?"
		args = append(args, *req.Description)
	}
	if req.Start != nil {
		query += ", start = ?"
		args = append(args, *req.Start)
	}
	if req.Duration != nil {
		query += ", duration = ?"
		args = append(args, *req.Duration)
	}
	if req.Canceled != nil {
		query += ", canceled = ?"
		args = append(args, *req.Canceled)
	}

	query += " WHERE event_id = ?"
	args = append(args, eventID)

	_, err = common.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventUpdate,
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUpdateEvent,
	})
}

// Delete supprime un événement (soft delete)
func (CalendarEventStruct) Delete(c *gin.Context) {
	_ = c.MustGet("user").(common.User)
	_ = c.MustGet("calendar").(common.Calendar)
	id := c.Param("event_id")
	eventID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidEventID,
		})
		return
	}

	// Vérifier si l'événement existe
	var existingEvent common.Event
	err = common.DB.QueryRow("SELECT event_id FROM event WHERE event_id = ? AND deleted_at IS NULL", eventID).Scan(&existingEvent.EventID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventNotFound,
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

	// Soft delete de l'événement
	_, err = tx.Exec("UPDATE event SET deleted_at = NOW() WHERE event_id = ?", eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventDelete,
		})
		return
	}

	// Soft delete des liaisons calendar_event
	_, err = tx.Exec("UPDATE calendar_event SET deleted_at = NOW() WHERE event_id = ?", eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarEventDeleteLink,
		})
		return
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessDeleteEvent,
	})
}
