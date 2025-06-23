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
	router.GET(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { UserCalendar.Get(c) },
	)
	router.GET(
		"/user-calendar/:user_id",
		middleware.UserExistsMiddleware("user_id"),
		func(c *gin.Context) { UserCalendar.List(c) },
	)
	router.POST(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { UserCalendar.Add(c) },
	)
	router.PUT(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { UserCalendar.Update(c) },
	)
	router.DELETE(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { UserCalendar.Delete(c) },
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

func TestUserCalendarCRUD(t *testing.T) {
	router := setupTestRouter()
	var userID int
	var userID2 int
	var calendarID int
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
			Title:       "Calendrier Test UserCalendar",
			Description: common.StringPtr("Description du calendrier de test pour user_calendar"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/calendar/%d", userID), bytes.NewBuffer(jsonData))
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
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d/%d", userID2, calendarID), nil)
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
	})

	t.Run("Create User Calendar Link (doublon)", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d/%d", userID2, calendarID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
		}
	})

	t.Run("Get User Calendar Link", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/user-calendar/%d/%d", userID2, calendarID), nil)
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

	t.Run("Update User Calendar Link", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/user-calendar/%d/%d", userID2, calendarID), nil)
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

	t.Run("Delete User Calendar Link", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/user-calendar/%d/%d", userID2, calendarID), nil)
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

func TestUserCalendarList(t *testing.T) {
	router := setupTestRouter()
	var userID int
	uniqueEmail := fmt.Sprintf("usercalendar.list+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "UserCalendarList",
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

	// Créer un premier calendrier
	{
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Test 1",
			Description: common.StringPtr("Description du premier calendrier"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/calendar/%d", userID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
	}

	// Créer un deuxième calendrier
	{
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Test 2",
			Description: common.StringPtr("Description du deuxième calendrier"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/calendar/%d", userID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
	}

	t.Run("List User Calendars", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/user-calendar/%d", userID), nil)
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

		// Vérifier que la réponse contient une liste de calendriers
		if response.Data == nil {
			t.Errorf("Expected data to not be nil")
		}

		// Vérifier que nous avons bien 2 calendriers
		calendars, ok := response.Data.([]interface{})
		if !ok {
			t.Errorf("Expected data to be an array")
		}
		if len(calendars) != 2 {
			t.Errorf("Expected 2 calendars, got %d", len(calendars))
		}
	})

	t.Run("List User Calendars - User Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user-calendar/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("List User Calendars - Invalid User ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user-calendar/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestUserCalendarErrorCases(t *testing.T) {
	router := setupTestRouter()

	// Création d'un utilisateur et d'un calendrier de test pour les cas d'erreur
	uniqueEmail := fmt.Sprintf("usercalendar.error+%d@test.com", time.Now().UnixNano())
	var testUserID int
	var testCalendarID int
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
	{
		payload := common.CreateCalendarRequest{
			Title:       "Calendrier Test Error",
			Description: common.StringPtr("Description du calendrier de test pour erreurs"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/calendar/%d", testUserID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["calendar_id"]; ok {
				testCalendarID = int(id.(float64))
			}
		}
	}

	t.Run("User inexistant", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/user-calendar/99999/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Calendar inexistant", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d/99999", testUserID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("ID user invalide", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/user-calendar/abc/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("ID calendar invalide", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d/abc", testUserID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Doublon", func(t *testing.T) {
		// Créer la liaison une première fois
		req, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d/%d", testUserID, testCalendarID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// Deuxième tentative (doit échouer)
		req2, _ := http.NewRequest("POST", fmt.Sprintf("/user-calendar/%d/%d", testUserID, testCalendarID), nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		if w2.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, w2.Code)
		}
	})
}
