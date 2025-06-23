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

	"go-averroes/internal/middleware"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	router := testutils.SetupTestRouter()
	// Configuration des routes pour les tests calendrier avec middlewares
	router.GET("/calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { Calendar.Get(c) },
	)
	router.POST("/calendar/:user_id",
		middleware.UserExistsMiddleware("user_id"),
		func(c *gin.Context) { Calendar.Add(c) },
	)
	router.PUT("/calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { Calendar.Update(c) },
	)
	router.DELETE("/calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { Calendar.Delete(c) },
	)
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) }) // Pour créer un user
	return router
}

func TestCalendarCRUD(t *testing.T) {
	router := setupTestRouter()
	var calendarID int
	uniqueEmail := fmt.Sprintf("calendar.user+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	var userID int
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
		payload := map[string]interface{}{
			"title":       "Calendrier Test",
			"description": "Description du calendrier de test",
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/%d", userID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf(common.ErrJSONParsing, err)
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
		url := fmt.Sprintf("/calendar/%d/%d", userID, calendarID)
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf(common.ErrJSONParsing, err)
		}
		if !response.Success {
			t.Errorf("Expected success true, got false")
		}
	})

	t.Run("Update Calendar", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Calendrier Modifié",
			"description": "Description modifiée",
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/%d/%d", userID, calendarID)
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
			t.Errorf(common.ErrJSONParsing, err)
		}
		if !response.Success {
			t.Errorf("Expected success true, got false")
		}
	})

	t.Run("Delete Calendar", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d/%d", userID, calendarID)
		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf(common.ErrJSONParsing, err)
		}
		if !response.Success {
			t.Errorf("Expected success true, got false")
		}
	})
}

func TestCalendarErrorCases(t *testing.T) {
	router := setupTestRouter()
	var calendarID int
	uniqueEmail := fmt.Sprintf("error.calendar.user+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests d'erreur
	var userID int
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
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["user_id"]; ok {
				userID = int(id.(float64))
			}
		}
	}

	// Créer un calendrier pour les tests d'erreur
	{
		payload := map[string]interface{}{
			"title":       "Calendrier Test Error",
			"description": "Description du calendrier de test pour erreurs",
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

	t.Run("Get Non-existent Calendar", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d/999999", userID)
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Get Calendar with Non-existent User", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/999999/%d", calendarID)
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create Calendar with Non-existent User", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Titre",
			"description": "Desc",
		}
		jsonData, _ := json.Marshal(payload)
		url := "/calendar/999999"
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update Calendar with Non-existent Calendar", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Titre",
			"description": "Desc",
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/%d/999999", userID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Update Calendar with Non-existent User", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Titre",
			"description": "Desc",
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/999999/%d", calendarID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete Calendar with Non-existent Calendar", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d/999999", userID)
		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Delete Calendar with Non-existent User", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/999999/%d", calendarID)
		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create Calendar with Invalid UserID", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Titre",
			"description": "Desc",
		}
		jsonData, _ := json.Marshal(payload)
		url := "/calendar/abc"
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Get Calendar with Invalid IDs", func(t *testing.T) {
		url := "/calendar/abc/def"
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
