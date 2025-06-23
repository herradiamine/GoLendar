package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
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

	// Configuration des routes pour les tests utilisateur
	router.GET("/user/:id", middleware.UserExistsMiddleware("id"), func(c *gin.Context) { User.Get(c) })
	router.POST("/user", func(c *gin.Context) { User.Add(c) })
	router.PUT("/user/:id", middleware.UserExistsMiddleware("id"), func(c *gin.Context) { User.Update(c) })
	router.DELETE("/user/:id", middleware.UserExistsMiddleware("id"), func(c *gin.Context) { User.Delete(c) })

	return router
}

func TestUserCRUD(t *testing.T) {
	router := setupTestRouter()
	var userID int
	uniqueEmail := fmt.Sprintf("jean.dupont+%d@test.com", time.Now().UnixNano())

	t.Run("Create User", func(t *testing.T) {
		payload := common.CreateUserRequest{
			Lastname:  "Dupont",
			Firstname: "Jean",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
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
		// Récupérer l'user_id créé
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["user_id"]; ok {
				userID = int(id.(float64))
			}
		}
	})

	t.Run("Get User", func(t *testing.T) {
		url := "/user/" + itoa(userID)
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

	t.Run("Update User", func(t *testing.T) {
		payload := common.UpdateUserRequest{
			Lastname:  stringPtr("Martin"),
			Firstname: stringPtr("Pierre"),
		}
		jsonData, _ := json.Marshal(payload)
		url := "/user/" + itoa(userID)
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

	t.Run("Delete User", func(t *testing.T) {
		url := "/user/" + itoa(userID)
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

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func TestUserErrorCases(t *testing.T) {
	router := setupTestRouter()
	var userID int
	uniqueEmail := fmt.Sprintf("error.user+%d@test.com", time.Now().UnixNano())

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
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["user_id"]; ok {
				userID = int(id.(float64))
			}
		}
	}

	t.Run("Get Non-existent User", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Create User with Invalid Email", func(t *testing.T) {
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "User",
			Email:     "invalid-email",
			Password:  "motdepasse123",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create User with Short Password", func(t *testing.T) {
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "User",
			Email:     "test@example.com",
			Password:  "123",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create User with Missing Required Fields", func(t *testing.T) {
		payload := map[string]interface{}{
			"lastname": "Test",
			// firstname manquant
			"email":    "test@example.com",
			"password": "motdepasse123",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Create User with Empty Fields", func(t *testing.T) {
		payload := common.CreateUserRequest{
			Lastname:  "",
			Firstname: "",
			Email:     "",
			Password:  "",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Update User with Invalid Email", func(t *testing.T) {
		payload := map[string]interface{}{
			"email": "invalid-email-format",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/user/%d", userID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Update User with Short Password", func(t *testing.T) {
		payload := map[string]interface{}{
			"password": "123",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/user/%d", userID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Update User with Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/user/%d", userID), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Get User with Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Update User with Invalid ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"lastname": "Updated",
		}

		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/user/invalid", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Delete User with Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/user/invalid", nil)
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
