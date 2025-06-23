package user_calendar

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
	// Configuration des routes pour les tests user_calendar
	router.GET("/user-calendar/:user_id/:user_calendar_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { UserCalendar.Get(c) })
	router.POST("/user-calendar/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { UserCalendar.Add(c) })
	router.PUT("/user-calendar/:user_id/:user_calendar_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { UserCalendar.Update(c) })
	router.DELETE("/user-calendar/:user_id/:user_calendar_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { UserCalendar.Delete(c) })
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })             // Pour créer un user
	router.POST("/calendar", func(c *gin.Context) { calendar.Calendar.Add(c) }) // Pour créer un calendar
	return router
}

func TestUserCalendarCRUD(t *testing.T) {
	router := setupTestRouter()
	var userID int
	var userID2 int
	var calendarID int
	var userCalendarID int
	uniqueEmail := fmt.Sprintf("usercalendar.user+%d@test.com", time.Now().UnixNano())
	uniqueEmail2 := fmt.Sprintf("usercalendar.user2+%d@test.com", time.Now().UnixNano())

	// Créer un premier utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "UserCalendar",
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

	// Créer un deuxième utilisateur pour la liaison user_calendar
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test2",
			Firstname: "UserCalendar2",
			Email:     uniqueEmail2,
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
				userID2 = int(id.(float64))
			}
		}
	}

	// Créer un calendrier pour les tests
	{
		payload := common.CreateCalendarRequest{
			UserID:      userID,
			Title:       "Calendrier Test UserCalendar",
			Description: stringPtr("Description du calendrier de test pour user_calendar"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
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

	t.Run("Create User Calendar Link", func(t *testing.T) {
		payload := map[string]interface{}{
			"calendar_id": calendarID,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d", userID2), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
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
			if id, ok := data["user_calendar_id"]; ok {
				userCalendarID = int(id.(float64))
			}
		}
	})

	t.Run("Get User Calendar Link", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/user-calendar/%d/%d", userID2, userCalendarID), nil)
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

	t.Run("Update User Calendar Link", func(t *testing.T) {
		payload := map[string]interface{}{
			"calendar_id": calendarID,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/user-calendar/%d/%d", userID2, userCalendarID), bytes.NewBuffer(jsonData))
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

	t.Run("Delete User Calendar Link", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/user-calendar/%d/%d", userID2, userCalendarID), nil)
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

func TestUserCalendarErrorCases(t *testing.T) {
	router := setupTestRouter()

	// Création d'un utilisateur de test pour les cas d'erreur
	uniqueEmail := fmt.Sprintf("usercalendar.error+%d@test.com", time.Now().UnixNano())
	var testUserID int
	{
		payload := common.CreateUserRequest{
			Lastname:  "Error",
			Firstname: "UserCalendar",
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
				testUserID = int(id.(float64))
			}
		}
	}

	t.Run("Get Non-existent User Calendar Link", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user-calendar/999/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create User Calendar Link with Non-existent User", func(t *testing.T) {
		payload := map[string]interface{}{
			"calendar_id": 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user-calendar/999", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create User Calendar Link with Non-existent Calendar", func(t *testing.T) {
		payload := map[string]interface{}{
			"calendar_id": 999,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d", testUserID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create User Calendar Link with Missing Required Fields", func(t *testing.T) {
		payload := map[string]interface{}{
			// calendar_id manquant
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d", testUserID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create User Calendar Link with Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d", testUserID), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Update User Calendar Link with Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/user-calendar/%d/%d", testUserID, 9999), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Get User Calendar Link with Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/user-calendar/%d/%d", testUserID, 9999), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update User Calendar Link with Invalid ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"calendar_id": 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/user-calendar/%d/%d", testUserID, 9999), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete User Calendar Link with Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/user-calendar/%d/%d", testUserID, 9999), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create User Calendar Link with Negative User ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"calendar_id": 1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user-calendar/-1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create User Calendar Link with Negative Calendar ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"calendar_id": -1,
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d", testUserID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}
