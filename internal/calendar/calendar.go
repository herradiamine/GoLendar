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
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    calendarData,
	})
}

// Add crée un nouveau calendrier pour un utilisateur
func (CalendarStruct) Add(c *gin.Context) {
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

	var req common.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
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
			Error:   common.ErrCalendarCreation,
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
			Error:   common.ErrUserCalendarLinkCreation,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessCreateCalendar,
		Data: gin.H{
			"calendar_id": calendarID,
			"user_id":     userID,
		},
	})
}

// Update met à jour un calendrier par user_id et calendar_id
func (CalendarStruct) Update(c *gin.Context) {
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}
	calendarID := calendarData.CalendarID

	var req common.UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
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
			Error:   common.ErrCalendarUpdate,
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUpdateCalendar,
	})
}

// Delete supprime un calendrier par user_id et calendar_id
func (CalendarStruct) Delete(c *gin.Context) {
	if _, ok := common.GetUserFromContext(c); !ok {
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
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	// Soft delete du calendrier
	_, err = tx.Exec("UPDATE calendar SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarDelete,
		})
		return
	}

	// Soft delete des liaisons user_calendar
	_, err = tx.Exec("UPDATE user_calendar SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarDeleteLink,
		})
		return
	}

	// Soft delete des liaisons calendar_event
	_, err = tx.Exec("UPDATE calendar_event SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarEventDeleteLink,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessDeleteCalendar,
	})
}
