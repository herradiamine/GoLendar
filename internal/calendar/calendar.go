// Package calendar internal/calendar/calendar.go
package calendar

import (
	"go-averroes/internal/common"
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"
)

type CalendarStruct struct{}

var Calendar = CalendarStruct{}

// Get récupère un calendrier par user_id et calendar_id
func (CalendarStruct) Get(c *gin.Context) {
	slog.Info(common.LogCalendarGet)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		slog.Error(common.LogCalendarGet + " - calendrier non trouvé dans le contexte")
		return
	}

	slog.Info(common.LogCalendarGet + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    calendarData,
	})
}

// Add crée un nouveau calendrier pour un utilisateur
func (CalendarStruct) Add(c *gin.Context) {
	slog.Info(common.LogCalendarAdd)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

	var req common.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(common.LogCalendarAdd + " - données invalides : " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogCalendarAdd + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
        INSERT INTO calendar (title, description, created_at) 
        VALUES (?, ?, NOW())
    `, req.Title, req.Description)
	if err != nil {
		slog.Error(common.LogCalendarAdd + " - erreur lors de la création du calendrier : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarCreation,
		})
		return
	}

	calendarID, _ := result.LastInsertId()

	_, err = tx.Exec(`
        INSERT INTO user_calendar (user_id, calendar_id, created_at) 
        VALUES (?, ?, NOW())
    `, userID, calendarID)
	if err != nil {
		slog.Error(common.LogCalendarAdd + " - erreur lors de la création de la liaison user_calendar : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarLinkCreation,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error(common.LogCalendarAdd + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogCalendarAdd + " - succès")
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
	slog.Info(common.LogCalendarUpdate)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		slog.Error(common.LogCalendarUpdate + " - calendrier non trouvé dans le contexte")
		return
	}
	calendarID := calendarData.CalendarID

	var req common.UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(common.LogCalendarUpdate + " - données invalides : " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

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
		slog.Error(common.LogCalendarUpdate + " - erreur lors de la mise à jour du calendrier : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarUpdate,
		})
		return
	}

	slog.Info(common.LogCalendarUpdate + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUpdateCalendar,
	})
}

// Delete supprime un calendrier par user_id et calendar_id
func (CalendarStruct) Delete(c *gin.Context) {
	slog.Info(common.LogCalendarDelete)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		slog.Error(common.LogCalendarDelete + " - calendrier non trouvé dans le contexte")
		return
	}
	calendarID := calendarData.CalendarID

	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogCalendarDelete + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE calendar SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		slog.Error(common.LogCalendarDelete + " - erreur lors de la suppression du calendrier : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarDelete,
		})
		return
	}

	_, err = tx.Exec("UPDATE user_calendar SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		slog.Error(common.LogCalendarDelete + " - erreur lors de la suppression des liaisons user_calendar : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCalendarDeleteLink,
		})
		return
	}

	_, err = tx.Exec("UPDATE calendar_event SET deleted_at = NOW() WHERE calendar_id = ?", calendarID)
	if err != nil {
		slog.Error(common.LogCalendarDelete + " - erreur lors de la suppression des liaisons calendar_event : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarEventDeleteLink,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error(common.LogCalendarDelete + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogCalendarDelete + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessDeleteCalendar,
	})
}
