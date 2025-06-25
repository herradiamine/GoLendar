package calendar_event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/calendar"
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/internal/session"
	"go-averroes/internal/user"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	router := testutils.SetupTestRouter()

	// Configuration des routes pour les tests événements calendrier avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })

	// Routes protégées par authentification
	router.POST("/calendar", middleware.AuthMiddleware(), func(c *gin.Context) { calendar.Calendar.Add(c) })
	router.GET("/calendar/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { calendar.Calendar.Get(c) })

	// Routes événements avec authentification et accès au calendrier
	router.GET("/calendar-event/:calendar_id/:event_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), middleware.EventExistsMiddleware("event_id"), func(c *gin.Context) { CalendarEvent.Get(c) })
	router.GET("/calendar-event/:calendar_id/month/:year/:month", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.ListByMonth(c) })
	router.GET("/calendar-event/:calendar_id/week/:year/:week", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.ListByWeek(c) })
	router.GET("/calendar-event/:calendar_id/day/:year/:month/:day", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.ListByDay(c) })
	router.POST("/calendar-event/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.Add(c) })
	router.PUT("/calendar-event/:calendar_id/:event_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), middleware.EventExistsMiddleware("event_id"), func(c *gin.Context) { CalendarEvent.Update(c) })
	router.DELETE("/calendar-event/:calendar_id/:event_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), middleware.EventExistsMiddleware("event_id"), func(c *gin.Context) { CalendarEvent.Delete(c) })

	return router
}

func TestCalendarEventCRUD(t *testing.T) {
	router := setupTestRouter()
	var calendarID int
	var eventID int
	var userToken string
	uniqueEmail := fmt.Sprintf("event.user+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Event",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Login pour obtenir un token
	{
		payload := common.LoginRequest{
			Email:    uniqueEmail,
			Password: "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if token, ok := data["session_token"]; ok {
				userToken = token.(string)
			}
		}
	}

	// Créer un calendrier pour les tests
	{
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Événements",
			Description: common.StringPtr("Calendrier pour tester les événements"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["calendar_id"]; ok {
				calendarID = int(id.(float64))
			}
		}
	}

	t.Run("Create Calendar Event", func(t *testing.T) {
		startTime := time.Now().Add(24 * time.Hour) // Demain
		payload := common.CreateEventRequest{
			Title:       "Événement Test",
			Description: common.StringPtr("Description de l'événement de test"),
			Start:       startTime,
			Duration:    60, // 60 minutes
			CalendarID:  calendarID,
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar-event/%d", calendarID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessCreateEvent, response.Message)

		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["event_id"]; ok {
				eventID = int(id.(float64))
			}
		}
	})

	t.Run("Get Calendar Event", func(t *testing.T) {
		url := fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Get Events By Month", func(t *testing.T) {
		now := time.Now()
		url := fmt.Sprintf("/calendar-event/%d/month/%d/%d", calendarID, now.Year(), now.Month())
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Get Events By Week", func(t *testing.T) {
		now := time.Now()
		year, week := now.ISOWeek()
		url := fmt.Sprintf("/calendar-event/%d/week/%d/%d", calendarID, year, week)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Get Events By Day", func(t *testing.T) {
		now := time.Now()
		url := fmt.Sprintf("/calendar-event/%d/day/%d/%d/%d", calendarID, now.Year(), now.Month(), now.Day())
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Update Calendar Event", func(t *testing.T) {
		payload := common.UpdateEventRequest{
			Title:       common.StringPtr("Événement Modifié"),
			Description: common.StringPtr("Description modifiée"),
			Duration:    common.IntPtr(120), // 2 heures
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessUpdateEvent, response.Message)
	})

	t.Run("Delete Calendar Event", func(t *testing.T) {
		url := fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessDeleteEvent, response.Message)
	})
}

func TestCalendarEventErrorCases(t *testing.T) {
	router := setupTestRouter()
	var calendarID int
	var eventID int
	var userToken string
	var otherUserToken string
	uniqueEmail := fmt.Sprintf("error.event.user+%d@test.com", time.Now().UnixNano())
	otherEmail := fmt.Sprintf("other.event.user+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests d'erreur
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Error",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Créer un autre utilisateur pour tester l'accès interdit
	{
		payload := common.CreateUserRequest{
			Lastname:  "Other",
			Firstname: "User",
			Email:     otherEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Login pour obtenir les tokens
	{
		payload := common.LoginRequest{
			Email:    uniqueEmail,
			Password: "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if token, ok := data["session_token"]; ok {
				userToken = token.(string)
			}
		}
	}

	{
		payload := common.LoginRequest{
			Email:    otherEmail,
			Password: "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if token, ok := data["session_token"]; ok {
				otherUserToken = token.(string)
			}
		}
	}

	// Créer un calendrier pour les tests d'erreur
	{
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Événements Error",
			Description: common.StringPtr("Calendrier pour tester les erreurs d'événements"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["calendar_id"]; ok {
				calendarID = int(id.(float64))
			}
		}
	}

	// Créer un événement pour les tests d'erreur
	{
		startTime := time.Now().Add(24 * time.Hour)
		payload := common.CreateEventRequest{
			Title:       "Événement Test Error",
			Description: common.StringPtr("Description de l'événement de test pour erreurs"),
			Start:       startTime,
			Duration:    60,
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar-event/%d", calendarID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["event_id"]; ok {
				eventID = int(id.(float64))
			}
		}
	}

	t.Run("Create Event Without Token", func(t *testing.T) {
		startTime := time.Now().Add(24 * time.Hour)
		payload := common.CreateEventRequest{
			Title:       "Événement Test",
			Description: common.StringPtr("Description de l'événement de test"),
			Start:       startTime,
			Duration:    60,
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar-event/%d", calendarID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotAuthenticated, response.Error)
	})

	t.Run("Access Event From Another User", func(t *testing.T) {
		url := fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+otherUserToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrNoAccessToCalendar, response.Error)
	})

	t.Run("Update Event From Another User", func(t *testing.T) {
		payload := common.UpdateEventRequest{
			Title: common.StringPtr("Événement Hacké"),
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+otherUserToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrNoAccessToCalendar, response.Error)
	})

	t.Run("Delete Event From Another User", func(t *testing.T) {
		url := fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", "Bearer "+otherUserToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrNoAccessToCalendar, response.Error)
	})

	t.Run("Get Non-existent Event", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/999", calendarID), nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrEventNotFound, response.Error)
	})

	t.Run("Create Event with Invalid Data", func(t *testing.T) {
		payload := map[string]interface{}{
			// Title manquant
			"description": "Description sans titre",
			"start":       time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04:05Z"),
			"duration":    60,
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar-event/%d", calendarID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Contains(t, response.Error, common.ErrInvalidData)
	})

	t.Run("Create Event with Invalid Date", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Événement Test",
			"description": "Description de l'événement de test",
			"start":       "date-invalide",
			"duration":    60,
			"calendar_id": calendarID,
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar-event/%d", calendarID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Contains(t, response.Error, common.ErrInvalidData)
	})
}

func TestParseDateFilter_Day(t *testing.T) {
	start, end, err := parseDateFilter("day", "2024-01-15")
	if err != nil {
		t.Errorf("parseDateFilter(day) erreur inattendue: %v", err)
	}
	if !start.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Début incorrect: %v", start)
	}
	if !end.Equal(time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Fin incorrecte: %v", end)
	}
}

func TestParseDateFilter_Week(t *testing.T) {
	start, end, err := parseDateFilter("week", "2024-W01")
	if err != nil {
		t.Errorf("parseDateFilter(week) erreur inattendue: %v", err)
	}
	if !start.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Début incorrect: %v", start)
	}
	if !end.Equal(time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Fin incorrecte: %v", end)
	}
}

func TestParseDateFilter_Month(t *testing.T) {
	start, end, err := parseDateFilter("month", "2024-01")
	if err != nil {
		t.Errorf("parseDateFilter(month) erreur inattendue: %v", err)
	}
	if !start.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Début incorrect: %v", start)
	}
	if !end.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Fin incorrecte: %v", end)
	}
}

func TestParseDateFilter_Invalid(t *testing.T) {
	_, _, err := parseDateFilter("day", "invalid-date")
	if err == nil {
		t.Error("Attendu une erreur pour une date invalide")
	}
	_, _, err = parseDateFilter("week", "2024-WXX")
	if err == nil {
		t.Error("Attendu une erreur pour une semaine invalide")
	}
	_, _, err = parseDateFilter("month", "2024-XX")
	if err == nil {
		t.Error("Attendu une erreur pour un mois invalide")
	}
	_, _, err = parseDateFilter("unknown", "2024-01-01")
	if err == nil {
		t.Error("Attendu une erreur pour un type de filtre inconnu")
	}
}
