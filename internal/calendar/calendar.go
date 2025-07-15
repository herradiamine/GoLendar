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
// @Summary Récupérer un calendrier
// @Description Récupère les informations d'un calendrier par son ID
// @Tags Calendrier
// @Produce json
// @Param calendar_id path int true "ID du calendrier"
// @Success 200 {object} common.JSONResponse
// @Failure 401 {object} common.JSONResponse
// @Failure 404 {object} common.JSONResponse
// @Router /calendar/{calendar_id} [get]
func (CalendarStruct) Get(c *gin.Context) {
	slog.Info(common.LogCalendarGet)
	if _, ok := common.GetUserFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotAuthenticated,
		})
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		slog.Error(common.LogCalendarGet + " - calendrier non trouvé dans le contexte")
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarNotFound,
		})
		return
	}

	slog.Info(common.LogCalendarGet + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessGetCalendar,
		Data:    calendarData,
	})
}

// Add crée un nouveau calendrier
// @Summary Créer un calendrier
// @Description Crée un nouveau calendrier
// @Tags Calendrier
// @Accept json
// @Produce json
// @Param calendrier body common.Calendar true "Données du calendrier"
// @Success 201 {object} common.JSONResponse
// @Failure 400 {object} common.JSONResponse
// @Failure 401 {object} common.JSONResponse
// @Router /calendar [post]
func (CalendarStruct) Add(c *gin.Context) {
	slog.Info(common.LogCalendarAdd)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		return
	}
	userID := userData.UserID

	slog.Info("Calendar.Add: Utilisateur récupéré", "user_id", userID)

	var req common.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(common.LogCalendarAdd + " - données invalides : " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	slog.Info("Calendar.Add: Données reçues", "title", req.Title, "description", req.Description)

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
	slog.Info("Calendar.Add: Calendrier créé", "calendar_id", calendarID)

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

	slog.Info("Calendar.Add: Liaison user_calendar créée", "user_id", userID, "calendar_id", calendarID)

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
// @Summary Mettre à jour un calendrier
// @Description Met à jour les informations d'un calendrier existant
// @Tags Calendrier
// @Accept json
// @Produce json
// @Param calendar_id path int true "ID du calendrier"
// @Param calendrier body common.Calendar true "Données du calendrier"
// @Success 200 {object} common.JSONResponse
// @Failure 400 {object} common.JSONResponse
// @Failure 401 {object} common.JSONResponse
// @Failure 404 {object} common.JSONResponse
// @Router /calendar/{calendar_id} [put]
func (CalendarStruct) Update(c *gin.Context) {
	slog.Info(common.LogCalendarUpdate)
	if _, ok := common.GetUserFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotAuthenticated,
		})
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		slog.Error(common.LogCalendarUpdate + " - calendrier non trouvé dans le contexte")
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarNotFound,
		})
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

	// Validation du titre obligatoire (non nil et non vide)
	if req.Title == nil || *req.Title == "" {
		slog.Error(common.LogCalendarUpdate + " - titre manquant ou vide")
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	query := "UPDATE calendar SET updated_at = NOW(), title = ?"
	args := []interface{}{*req.Title}
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
// @Summary Supprimer un calendrier
// @Description Supprime un calendrier par son ID
// @Tags Calendrier
// @Produce json
// @Param calendar_id path int true "ID du calendrier"
// @Success 204 {object} common.JSONResponse
// @Failure 401 {object} common.JSONResponse
// @Failure 404 {object} common.JSONResponse
// @Router /calendar/{calendar_id} [delete]
func (CalendarStruct) Delete(c *gin.Context) {
	slog.Info(common.LogCalendarDelete)
	if _, ok := common.GetUserFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotAuthenticated,
		})
		return
	}

	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		slog.Error(common.LogCalendarDelete + " - calendrier non trouvé dans le contexte")
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarNotFound,
		})
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
