// Package calendar_event internal/calendar_event/calendar_event.go
package calendar_event

import (
	"fmt"
	"go-averroes/internal/common"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CalendarEventStruct struct{}

var CalendarEvent = CalendarEventStruct{}

// Get récupère un événement par son ID
func (CalendarEventStruct) Get(c *gin.Context) {
	slog.Info(common.LogEventGet)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}
	if _, ok := common.GetCalendarFromContext(c); !ok {
		return
	}

	eventData, ok := common.GetEventFromContext(c)
	if !ok {
		slog.Error(common.LogEventGet + " - événement non trouvé dans le contexte")
		return
	}

	slog.Info(common.LogEventGet + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    eventData,
	})
}

// Add crée un nouvel événement
func (CalendarEventStruct) Add(c *gin.Context) {
	slog.Info(common.LogEventAdd)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}
	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}
	calendarID := calendarData.CalendarID

	var req common.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(common.LogEventAdd + " - données invalides : " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	if req.Duration < 1 {
		slog.Error(common.LogEventAdd + " - durée invalide")
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidDuration,
		})
		return
	}

	// La vérification d'accès est maintenant gérée par le middleware UserCanAccessCalendarMiddleware

	// Valeur par défaut pour canceled si non fournie
	canceled := false
	if req.Canceled != nil {
		canceled = *req.Canceled
	}

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogEventAdd + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
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
		slog.Error(common.LogEventAdd + " - erreur lors de la création de l'événement : " + err.Error())
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
		slog.Error(common.LogEventAdd + " - erreur lors de la création de la liaison calendar_event : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarEventLink,
		})
		return
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		slog.Error(common.LogEventAdd + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogEventAdd + " - succès")
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
	slog.Info(common.LogEventUpdate)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}
	if _, ok := common.GetCalendarFromContext(c); !ok {
		return
	}

	eventData, ok := common.GetEventFromContext(c)
	if !ok {
		slog.Error(common.LogEventUpdate + " - événement non trouvé dans le contexte")
		return
	}
	eventID := eventData.EventID

	var req common.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(common.LogEventUpdate + " - données invalides : " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	if req.Duration != nil && *req.Duration < 1 {
		slog.Error(common.LogEventUpdate + " - durée invalide")
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidDuration,
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

	_, err := common.DB.Exec(query, args...)
	if err != nil {
		slog.Error(common.LogEventUpdate + " - erreur lors de la mise à jour de l'événement : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventUpdate,
		})
		return
	}

	slog.Info(common.LogEventUpdate + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUpdateEvent,
	})
}

// Delete supprime un événement (soft delete)
func (CalendarEventStruct) Delete(c *gin.Context) {
	slog.Info(common.LogEventDelete)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}
	if _, ok := common.GetCalendarFromContext(c); !ok {
		return
	}

	eventData, ok := common.GetEventFromContext(c)
	if !ok {
		slog.Error(common.LogEventDelete + " - événement non trouvé dans le contexte")
		return
	}
	eventID := eventData.EventID

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogEventDelete + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Soft delete de l'événement
	_, err = tx.Exec("UPDATE event SET deleted_at = NOW() WHERE event_id = ?", eventID)
	if err != nil {
		slog.Error(common.LogEventDelete + " - erreur lors de la suppression de l'événement : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventDelete,
		})
		return
	}

	// Soft delete des liaisons calendar_event
	_, err = tx.Exec("UPDATE calendar_event SET deleted_at = NOW() WHERE event_id = ?", eventID)
	if err != nil {
		slog.Error(common.LogEventDelete + " - erreur lors de la suppression de la liaison calendar_event : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrCalendarEventDeleteLink,
		})
		return
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		slog.Error(common.LogEventDelete + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogEventDelete + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessDeleteEvent,
	})
}

// ListByMonth récupère les événements d'un calendrier pour un mois donné
func (CalendarEventStruct) ListByMonth(c *gin.Context) {
	yearStr := c.Param("year")
	monthStr := c.Param("month")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{Success: false, Error: common.ErrInvalidYear})
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{Success: false, Error: common.ErrInvalidMonth})
		return
	}
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	listEventsWithRange(c, startDate, endDate)
}

// ListByWeek récupère les événements d'un calendrier pour une semaine ISO donnée
func (CalendarEventStruct) ListByWeek(c *gin.Context) {
	yearStr := c.Param("year")
	weekStr := c.Param("week")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{Success: false, Error: common.ErrInvalidYear})
		return
	}
	week, err := strconv.Atoi(weekStr)
	if err != nil || week < 1 || week > 53 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{Success: false, Error: common.ErrInvalidWeekNumber})
		return
	}
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	week1Monday := jan4.AddDate(0, 0, 1-weekday)
	startDate := week1Monday.AddDate(0, 0, (week-1)*7)
	endDate := startDate.AddDate(0, 0, 7)
	listEventsWithRange(c, startDate, endDate)
}

// ListByDay récupère les événements d'un calendrier pour un jour donné
func (CalendarEventStruct) ListByDay(c *gin.Context) {
	yearStr := c.Param("year")
	monthStr := c.Param("month")
	dayStr := c.Param("day")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{Success: false, Error: common.ErrInvalidYear})
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{Success: false, Error: common.ErrInvalidMonth})
		return
	}
	day, err := strconv.Atoi(dayStr)
	if err != nil || day < 1 || day > 31 {
		c.JSON(http.StatusBadRequest, common.JSONResponse{Success: false, Error: common.ErrInvalidDay})
		return
	}
	startDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 1)
	listEventsWithRange(c, startDate, endDate)
}

// listEventsWithRange est une fonction utilitaire pour factoriser la logique de récupération
func listEventsWithRange(c *gin.Context, startDate, endDate time.Time) {
	slog.Info(common.LogEventList)
	if _, ok := common.GetUserFromContext(c); !ok {
		return
	}
	calendarData, ok := common.GetCalendarFromContext(c)
	if !ok {
		return
	}
	calendarID := calendarData.CalendarID

	query := `
		SELECT e.event_id, e.title, e.description, e.start, e.duration, e.canceled, 
		       e.created_at, e.updated_at, e.deleted_at 
		FROM event e
		INNER JOIN calendar_event ce ON e.event_id = ce.event_id
		WHERE ce.calendar_id = ? AND ce.deleted_at IS NULL 
		  AND e.deleted_at IS NULL 
		  AND e.start >= ? AND e.start < ?
		ORDER BY e.start ASC
	`
	rows, err := common.DB.Query(query, calendarID, startDate, endDate)
	if err != nil {
		slog.Error(common.LogEventList + " - erreur lors de la récupération des événements : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventsRetrieval,
		})
		return
	}
	defer rows.Close()

	var events []common.Event
	for rows.Next() {
		var event common.Event
		err := rows.Scan(&event.EventID, &event.Title, &event.Description, &event.Start, &event.Duration, &event.Canceled, &event.CreatedAt, &event.UpdatedAt, &event.DeletedAt)
		if err != nil {
			slog.Error(common.LogEventList + " - erreur lors de la lecture des événements : " + err.Error())
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   common.ErrEventsReading,
			})
			return
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		slog.Error(common.LogEventList + " - erreur lors de l'itération des résultats : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrEventsRetrieval,
		})
		return
	}

	slog.Info(fmt.Sprintf("%s - succès, %d événements trouvés", common.LogEventList, len(events)))
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessListEvents,
		Data:    events,
	})
}
