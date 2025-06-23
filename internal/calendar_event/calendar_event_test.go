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
		middleware.UserCanAccessCalendarMiddleware(),
		func(c *gin.Context) { CalendarEvent.Get(c) },
	)
	router.GET(
		"/calendar-event/:user_id/:calendar_id/month/:year/:month",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		middleware.UserCanAccessCalendarMiddleware(),
		func(c *gin.Context) { CalendarEvent.ListByMonth(c) },
	)
	router.GET(
		"/calendar-event/:user_id/:calendar_id/week/:year/:week",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		middleware.UserCanAccessCalendarMiddleware(),
		func(c *gin.Context) { CalendarEvent.ListByWeek(c) },
	)
	router.GET(
		"/calendar-event/:user_id/:calendar_id/day/:year/:month/:day",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		middleware.UserCanAccessCalendarMiddleware(),
		func(c *gin.Context) { CalendarEvent.ListByDay(c) },
	)
	router.GET(
		"/calendar-event/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		middleware.UserCanAccessCalendarMiddleware(),
		func(c *gin.Context) { CalendarEvent.List(c) },
	)
	router.POST(
		"/calendar-event/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		middleware.UserCanAccessCalendarMiddleware(),
		func(c *gin.Context) { CalendarEvent.Add(c) },
	)
	router.PUT(
		"/calendar-event/:user_id/:calendar_id/:event_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		middleware.UserCanAccessCalendarMiddleware(),
		func(c *gin.Context) { CalendarEvent.Update(c) },
	)
	router.DELETE(
		"/calendar-event/:user_id/:calendar_id/:event_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		middleware.UserCanAccessCalendarMiddleware(),
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
			Description: common.StringPtr("Description du calendrier de test pour événements"),
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
			Description: common.StringPtr("Description de l'événement de test"),
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
			Title:    common.StringPtr("Événement Modifié"),
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

func TestEventList(t *testing.T) {
	router := setupTestRouter()
	var userID int
	var calendarID int
	uniqueEmail := fmt.Sprintf("event.list+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "EventList",
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
			Title:       "Calendrier Test EventList",
			Description: common.StringPtr("Description du calendrier de test pour liste d'événements"),
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

	// Créer plusieurs événements pour tester les filtres
	eventDates := []time.Time{
		parseTime("2024-01-15T10:00:00Z"), // Jour spécifique
		parseTime("2024-01-16T14:00:00Z"), // Même semaine
		parseTime("2024-01-20T09:00:00Z"), // Même mois
		parseTime("2024-02-01T16:00:00Z"), // Mois différent
	}

	for i, startTime := range eventDates {
		payload := common.CreateEventRequest{
			Title:       fmt.Sprintf("Événement Test %d", i+1),
			Description: common.StringPtr(fmt.Sprintf("Description de l'événement %d", i+1)),
			Start:       startTime,
			Duration:    60,
			CalendarID:  calendarID,
			Canceled:    boolPtr(false),
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/calendar-event/%d/%d", userID, calendarID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	t.Run("List Events by Day", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/%d/day/2024/1/15", userID, calendarID), nil)
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

		// Vérifier que nous avons bien 1 événement pour ce jour
		events, ok := response.Data.([]interface{})
		if !ok {
			t.Errorf("Expected data to be an array")
		}
		if len(events) != 1 {
			t.Errorf("Expected 1 event for the day, got %d", len(events))
		}
	})

	t.Run("List Events by Week", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/%d/week/2024/3", userID, calendarID), nil)
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

		// Vérifier que nous avons bien au moins 1 événement pour cette semaine
		events, ok := response.Data.([]interface{})
		if !ok {
			t.Errorf("Expected data to be an array")
		}
		if len(events) < 1 {
			t.Errorf("Expected at least 1 event for the week, got %d", len(events))
		}
	})

	t.Run("List Events by Month", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/%d/month/2024/1", userID, calendarID), nil)
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

		// Vérifier que nous avons bien 3 événements pour ce mois
		events, ok := response.Data.([]interface{})
		if !ok {
			t.Errorf("Expected data to be an array")
		}
		if len(events) != 3 {
			t.Errorf("Expected 3 events for the month, got %d", len(events))
		}
	})

	t.Run("List Events - Missing Parameters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/%d", userID, calendarID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("List Events - Invalid Filter Type", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/%d?filter_type=invalid&date=2024-01-15", userID, calendarID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("List Events - Invalid Date Format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/calendar-event/%d/%d?filter_type=day&date=invalid-date", userID, calendarID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
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
			Title: common.StringPtr("Événement Modifié"),
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

	t.Run("Create Event with No Access to Calendar", func(t *testing.T) {
		// Créer un utilisateur principal pour le test
		var mainUserID int
		{
			userPayload := common.CreateUserRequest{
				Lastname:  "Main",
				Firstname: "User",
				Email:     fmt.Sprintf("main.user.forbidden-%d@test.com", time.Now().UnixNano()),
				Password:  "password123",
			}
			jsonData, _ := json.Marshal(userPayload)
			req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			var resp common.JSONResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			data := resp.Data.(map[string]interface{})
			mainUserID = int(data["user_id"].(float64))
		}

		// Créer un autre utilisateur et son calendrier
		var otherCalendarID int
		{
			otherUserPayload := common.CreateUserRequest{
				Lastname:  "Other",
				Firstname: "User",
				Email:     fmt.Sprintf("other.user.forbidden-%d@test.com", time.Now().UnixNano()),
				Password:  "password123",
			}
			jsonData, _ := json.Marshal(otherUserPayload)
			req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			var resp common.JSONResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			data := resp.Data.(map[string]interface{})
			otherUserID := int(data["user_id"].(float64))

			calendarPayload := common.CreateCalendarRequest{Title: "Other's Calendar"}
			jsonCalData, _ := json.Marshal(calendarPayload)
			reqCal, _ := http.NewRequest("POST", fmt.Sprintf("/calendar/%d", otherUserID), bytes.NewBuffer(jsonCalData))
			reqCal.Header.Set("Content-Type", "application/json")
			wCal := httptest.NewRecorder()
			router.ServeHTTP(wCal, reqCal)
			var calResp common.JSONResponse
			json.Unmarshal(wCal.Body.Bytes(), &calResp)
			calData := calResp.Data.(map[string]interface{})
			otherCalendarID = int(calData["calendar_id"].(float64))
		}

		// Tenter de créer un événement dans le calendrier de l'autre utilisateur
		eventPayload := common.CreateEventRequest{
			Title:    "Unauthorized Event",
			Start:    time.Now(),
			Duration: 60,
		}
		jsonData, _ := json.Marshal(eventPayload)

		url := fmt.Sprintf("/calendar-event/%d/%d", mainUserID, otherCalendarID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
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
