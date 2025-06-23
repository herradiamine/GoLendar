package calendar_event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/calendar"
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/internal/user"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func setupTestRouter() *gin.Engine {
	router := testutils.SetupTestRouter()
	// Configuration des routes pour les tests calendar_event
	router.GET(
		"/calendar-event/:user_id/:calendar_id/:event_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { CalendarEvent.Get(c) },
	)
	router.POST(
		"/calendar-event/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { CalendarEvent.Add(c) },
	)
	router.PUT(
		"/calendar-event/:user_id/:calendar_id/:event_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { CalendarEvent.Update(c) },
	)
	router.DELETE(
		"/calendar-event/:user_id/:calendar_id/:event_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { CalendarEvent.Delete(c) },
	)
	router.POST(
		"/user",
		func(c *gin.Context) { user.User.Add(c) },
	) // Pour créer un user
	router.POST(
		"/calendar/:user_id",
		middleware.UserExistsMiddleware("user_id"),
		func(c *gin.Context) { calendar.Calendar.Add(c) },
	) // Pour créer un calendar
	return router
}

func TestEventCRUD(t *testing.T) {
	router := setupTestRouter()
	var userID int
	var calendarID int
	var eventID int
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
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["user_id"]; ok {
				userID = int(id.(float64))
			}
		}
	}

	// Créer un calendrier pour les tests
	{
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Test Event",
			Description: stringPtr("Description du calendrier de test pour événements"),
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/%d", userID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
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

	t.Run("Create Event", func(t *testing.T) {
		payload := common.CreateEventRequest{
			Title:       "Événement Test",
			Description: stringPtr("Description de l'événement de test"),
			Start:       parseTime("2024-01-15T10:00:00Z"),
			Duration:    60,
			CalendarID:  calendarID,
			Canceled:    boolPtr(false),
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/calendar-event/%d/%d", userID, calendarID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		fmt.Printf("Expected status %d, got %d\n", http.StatusCreated, w.Code)
		fmt.Printf("Body: %s\n", w.Body.String())
		if w.Code != http.StatusCreated {
		}

		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Erreur parsing JSON: %v", err)
		}
		if !response.Success {
			t.Errorf("Expected success true, got false")
		}
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["event_id"]; ok {
				eventID = int(id.(float64))
			}
		}
	})

	t.Run("Get Event", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/%d/%d", userID, calendarID, eventID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Erreur parsing JSON: %v", err)
		}
		if !response.Success {
			t.Errorf("Expected success true, got false")
		}
	})

	t.Run("Update Event", func(t *testing.T) {
		payload := common.UpdateEventRequest{
			Title:    stringPtr("Événement Modifié"),
			Duration: intPtr(90),
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/calendar-event/%d/%d/%d", userID, calendarID, eventID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Erreur parsing JSON: %v", err)
		}
		if !response.Success {
			t.Errorf("Expected success true, got false")
		}
	})

	t.Run("Delete Event", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/calendar-event/%d/%d/%d", userID, calendarID, eventID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Erreur parsing JSON: %v", err)
		}
		if !response.Success {
			t.Errorf("Expected success true, got false")
		}
	})
}

func TestEventErrorCases(t *testing.T) {
	router := setupTestRouter()

	t.Run("Get Non-existent Event", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/calendar-event/999/999/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create Event without User ID", func(t *testing.T) {
		payload := common.CreateEventRequest{
			Title:      "Test Event",
			Start:      parseTime("2024-01-15T10:00:00Z"),
			Duration:   60,
			CalendarID: 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar-event/99999/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Create Event with Non-existent Calendar", func(t *testing.T) {
		payload := common.CreateEventRequest{
			Title:      "Test Event",
			Start:      parseTime("2024-01-15T10:00:00Z"),
			Duration:   60,
			CalendarID: 999,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar-event/1/999", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Create Event with Missing Required Fields", func(t *testing.T) {
		payload := map[string]interface{}{
			// title manquant
			"start":       "2024-01-15T10:00:00Z",
			"duration":    60,
			"calendar_id": 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar-event/1/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Create Event with Invalid Duration", func(t *testing.T) {
		payload := common.CreateEventRequest{
			Title:      "Test Event",
			Start:      parseTime("2024-01-15T10:00:00Z"),
			Duration:   0, // Durée invalide (doit être >= 1)
			CalendarID: 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar-event/1/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Create Event with Negative Duration", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Test Event",
			"start":       "2024-01-15T10:00:00Z",
			"duration":    -10,
			"calendar_id": 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar-event/1/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Create Event with Empty Title", func(t *testing.T) {
		payload := common.CreateEventRequest{
			Title:      "",
			Start:      parseTime("2024-01-15T10:00:00Z"),
			Duration:   60,
			CalendarID: 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar-event/1/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Create Event with Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/calendar-event/1/1", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Update Event with Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/calendar-event/1/1/1", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Update Event with Invalid Duration", func(t *testing.T) {
		payload := map[string]interface{}{
			"duration": 0, // Durée invalide
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/calendar-event/1/1/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Get Event with Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/calendar-event/invalid/invalid/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Update Event with Invalid ID", func(t *testing.T) {
		payload := common.UpdateEventRequest{
			Title: stringPtr("Événement Modifié"),
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/calendar-event/invalid/invalid/invalid", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})

	t.Run("Delete Event with Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/calendar-event/invalid/invalid/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404, got %d", w.Code)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}
