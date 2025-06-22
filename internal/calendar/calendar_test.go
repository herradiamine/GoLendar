package calendar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/common"
	"go-averroes/internal/user"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	router := testutils.SetupTestRouter()
	// Configuration des routes pour les tests calendrier
	router.GET("/calendar/:id", func(c *gin.Context) { Calendar.Get(c) })
	router.POST("/calendar", func(c *gin.Context) { Calendar.Add(c) })
	router.PUT("/calendar/:id", func(c *gin.Context) { Calendar.Update(c) })
	router.DELETE("/calendar/:id", func(c *gin.Context) { Calendar.Delete(c) })
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) }) // Pour créer un user
	return router
}

func TestCalendarCRUD(t *testing.T) {
	router := setupTestRouter()
	var userID int
	var calendarID int
	uniqueEmail := fmt.Sprintf("calendar.user+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Calendar",
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

	t.Run("Create Calendar", func(t *testing.T) {
		payload := common.CreateCalendarRequest{
			UserID:      userID,
			Title:       "Calendrier Test",
			Description: stringPtr("Description du calendrier de test"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
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
			if id, ok := data["calendar_id"]; ok {
				calendarID = int(id.(float64))
			}
		}
	})

	t.Run("Get Calendar", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d", calendarID)
		req, _ := http.NewRequest("GET", url, nil)
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

	t.Run("Update Calendar", func(t *testing.T) {
		payload := common.UpdateCalendarRequest{
			Title:       stringPtr("Calendrier Modifié"),
			Description: stringPtr("Description modifiée"),
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/%d", calendarID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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

	t.Run("Delete Calendar", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d", calendarID)
		req, _ := http.NewRequest("DELETE", url, nil)
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

func TestCalendarErrorCases(t *testing.T) {
	router := setupTestRouter()

	t.Run("Get Non-existent Calendar", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/calendar/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create Calendar with Non-existent User", func(t *testing.T) {
		payload := common.CreateCalendarRequest{
			UserID:      999,
			Title:       "Calendrier Test",
			Description: stringPtr("Description du calendrier de test"),
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
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
