package calendar

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	// Configuration des routes pour les tests calendrier avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })

	// Routes protégées par authentification
	router.POST("/calendar", middleware.AuthMiddleware(), func(c *gin.Context) { Calendar.Add(c) })
	router.GET("/calendar/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { Calendar.Get(c) })
	router.PUT("/calendar/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { Calendar.Update(c) })
	router.DELETE("/calendar/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { Calendar.Delete(c) })

	return router
}

func TestCalendarCRUD(t *testing.T) {
	router := setupTestRouter()
	var calendarID int
	var userToken string
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

	t.Run("Create Calendar", func(t *testing.T) {
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Test",
			Description: common.StringPtr("Description du calendrier de test"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessCreateCalendar, response.Message)

		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["calendar_id"]; ok {
				calendarID = int(id.(float64))
			}
		}
	})

	t.Run("Get Calendar", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d", calendarID)
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

	t.Run("Update Calendar", func(t *testing.T) {
		payload := common.UpdateCalendarRequest{
			Title:       common.StringPtr("Calendrier Modifié"),
			Description: common.StringPtr("Description modifiée"),
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/%d", calendarID)
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
		require.Equal(t, common.MsgSuccessUpdateCalendar, response.Message)
	})

	t.Run("Delete Calendar", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d", calendarID)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessDeleteCalendar, response.Message)
	})
}

func TestCalendarErrorCases(t *testing.T) {
	router := setupTestRouter()
	var calendarID int
	var userToken string
	var otherUserToken string
	uniqueEmail := fmt.Sprintf("error.calendar.user+%d@test.com", time.Now().UnixNano())
	otherEmail := fmt.Sprintf("other.calendar.user+%d@test.com", time.Now().UnixNano())

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
			Title:       "Calendrier Test Error",
			Description: common.StringPtr("Description du calendrier de test pour erreurs"),
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

	t.Run("Create Calendar Without Token", func(t *testing.T) {
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Test",
			Description: common.StringPtr("Description du calendrier de test"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
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

	t.Run("Get Calendar Without Token", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d", calendarID)
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotAuthenticated, response.Error)
	})

	t.Run("Access Calendar From Another User", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d", calendarID)
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

	t.Run("Update Calendar From Another User", func(t *testing.T) {
		payload := common.UpdateCalendarRequest{
			Title: common.StringPtr("Calendrier Hacké"),
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/calendar/%d", calendarID)
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

	t.Run("Delete Calendar From Another User", func(t *testing.T) {
		url := fmt.Sprintf("/calendar/%d", calendarID)
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

	t.Run("Get Non-existent Calendar", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/calendar/999", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrCalendarNotFound, response.Error)
	})

	t.Run("Create Calendar with Invalid Data", func(t *testing.T) {
		payload := map[string]interface{}{
			// Title manquant
			"description": "Description sans titre",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/calendar", bytes.NewBuffer(jsonData))
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
