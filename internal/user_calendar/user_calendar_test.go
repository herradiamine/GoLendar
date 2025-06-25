package user_calendar

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

	// Configuration des routes pour les tests user_calendar avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })

	// Routes protégées par authentification
	router.POST("/calendar", middleware.AuthMiddleware(), func(c *gin.Context) { calendar.Calendar.Add(c) })
	router.GET("/calendar/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { calendar.Calendar.Get(c) })

	// Routes user_calendar avec authentification et accès au calendrier
	router.GET("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { UserCalendar.Get(c) })
	router.GET("/user-calendar/:user_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { UserCalendar.GetByUser(c) })
	router.POST("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) { UserCalendar.Add(c) })
	router.PUT("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) { UserCalendar.Update(c) })
	router.DELETE("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) { UserCalendar.Delete(c) })

	return router
}

func TestUserCalendarCRUD(t *testing.T) {
	router := setupTestRouter()
	var userID int
	var calendarID int
	var userToken string
	uniqueEmail := fmt.Sprintf("user.calendar+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
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
			Title:       "Calendrier User Calendar",
			Description: common.StringPtr("Calendrier pour tester les liaisons user-calendar"),
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

	t.Run("Create User Calendar Link", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Si c'est la première création, on attend 201, sinon 409 (conflit)
		if w.Code == http.StatusCreated {
			require.Equal(t, http.StatusCreated, w.Code)
			var response common.JSONResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			require.True(t, response.Success)
			require.Equal(t, common.MsgSuccessCreateUserCalendar, response.Message)
		} else {
			require.Equal(t, http.StatusConflict, w.Code)
			var response common.JSONResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			require.False(t, response.Success)
			require.Equal(t, common.ErrUserCalendarConflict, response.Error)
		}
	})

	t.Run("Get User Calendar Link", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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

	t.Run("Get User Calendars", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d", userID)
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

	t.Run("Update User Calendar Link", func(t *testing.T) {
		payload := map[string]interface{}{
			"permissions": "read_write",
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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
		require.Equal(t, common.MsgSuccessUpdateUserCalendar, response.Message)
	})

	t.Run("Delete User Calendar Link", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessDeleteUserCalendar, response.Message)
	})
}

func TestUserCalendarErrorCases(t *testing.T) {
	router := setupTestRouter()
	var userID int
	var calendarID int
	var userToken string
	var otherUserToken string
	uniqueEmail := fmt.Sprintf("error.user.calendar+%d@test.com", time.Now().UnixNano())
	otherEmail := fmt.Sprintf("other.user.calendar+%d@test.com", time.Now().UnixNano())

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
			Title:       "Calendrier User Calendar Error",
			Description: common.StringPtr("Calendrier pour tester les erreurs de liaisons user-calendar"),
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

	// Créer la liaison user-calendar pour le premier utilisateur
	{
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// On ignore le résultat, on veut juste créer la liaison
	}

	t.Run("Create User Calendar Link Without Token", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
		req, _ := http.NewRequest("POST", url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotAuthenticated, response.Error)
	})

	t.Run("Get User Calendar Link Without Token", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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

	t.Run("Access User Calendar From Another User", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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

	t.Run("Update User Calendar From Another User", func(t *testing.T) {
		payload := map[string]interface{}{
			"permissions": "admin",
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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

	t.Run("Delete User Calendar From Another User", func(t *testing.T) {
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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

	t.Run("Get Non-existent User Calendar", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/user-calendar/%d/999", userID), nil)
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

	t.Run("Get User Calendar with Non-existent User", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/user-calendar/999/%d", calendarID), nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotFound, response.Error)
	})

	t.Run("Create User Calendar with Invalid Data", func(t *testing.T) {
		payload := map[string]interface{}{
			"invalid_field": "invalid_value",
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Puisque la liaison existe déjà, on attend 409 (Conflict)
		require.Equal(t, http.StatusConflict, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserCalendarConflict, response.Error)
	})
}

func TestUserCalendar_List_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/user-calendar", func(c *gin.Context) {
		UserCalendar.List(c)
	})
	req, _ := http.NewRequest("GET", "/user-calendar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == 200 {
		t.Error("List devrait bloquer sans utilisateur dans le contexte")
	}
}

func TestUserCalendar_List_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth_user", common.User{UserID: 1})
		if common.DB != nil {
			_ = common.DB.Close()
		}
	})
	r.GET("/user-calendar", func(c *gin.Context) {
		UserCalendar.List(c)
	})
	req, _ := http.NewRequest("GET", "/user-calendar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Errorf("List avec DB fermée: code HTTP = %d, want 500 ou 400", w.Code)
	}
}

func TestAdd_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/user-calendar", func(c *gin.Context) {
		UserCalendar.Add(c)
	})
	req, _ := http.NewRequest("POST", "/user-calendar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == 200 {
		t.Error("Add devrait bloquer sans utilisateur dans le contexte")
	}
}

func TestAdd_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth_user", common.User{UserID: 1})
		if common.DB != nil {
			_ = common.DB.Close()
		}
	})
	r.POST("/user-calendar", func(c *gin.Context) {
		UserCalendar.Add(c)
	})
	req, _ := http.NewRequest("POST", "/user-calendar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Errorf("Add avec DB fermée: code HTTP = %d, want 500 ou 400", w.Code)
	}
}

func TestList_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth_user", common.User{UserID: 1})
		if common.DB != nil {
			_ = common.DB.Close()
		}
	})
	r.GET("/user-calendar", func(c *gin.Context) {
		UserCalendar.List(c)
	})
	req, _ := http.NewRequest("GET", "/user-calendar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Errorf("List avec DB fermée: code HTTP = %d, want 500 ou 400", w.Code)
	}
}
