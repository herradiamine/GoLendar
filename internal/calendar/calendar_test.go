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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

// --- Helpers mutualisés ---
// Retourne l'ID de l'utilisateur créé
func createUser(router http.Handler, email, password, firstname, lastname string) (int, *http.Response) {
	payload := map[string]string{
		"email":     email,
		"password":  password,
		"firstname": firstname,
		"lastname":  lastname,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Vérifier que la création a réussi
	if w.Code != http.StatusCreated {
		return 0, w.Result()
	}

	// Récupérer l'ID créé
	var userID int
	row := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", email)
	err := row.Scan(&userID)
	if err != nil {
		return 0, w.Result()
	}
	return userID, w.Result()
}

func loginAndGetToken(router http.Handler, email, password string) (string, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	var jsonResp struct {
		Success bool `json:"success"`
		Data    struct {
			SessionToken string `json:"session_token"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	if !jsonResp.Success || jsonResp.Data.SessionToken == "" {
		return "", fmt.Errorf("login failed: %s", jsonResp.Error)
	}
	return jsonResp.Data.SessionToken, nil
}

// Retourne l'ID du calendrier créé
func createCalendar(router http.Handler, token, title, description string) (int, *http.Response) {
	payload := map[string]string{
		"title":       title,
		"description": description,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/calendar", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Vérifier que la création a réussi
	if w.Code != http.StatusCreated {
		return 0, w.Result()
	}

	// Récupérer l'ID créé
	var calendarID int
	row := common.DB.QueryRow("SELECT calendar_id FROM calendar WHERE title = ? AND description = ?", title, description)
	err := row.Scan(&calendarID)
	if err != nil {
		return 0, w.Result()
	}
	return calendarID, w.Result()
}

// --- Test table-driven pour GET /calendar/:calendar_id ---
func TestGetCalendar(t *testing.T) {
	var TestCases = []struct {
		CaseName        string
		Prepare         func(router *gin.Engine) (token string, calendarID string, otherToken string)
		Token           string
		CalendarID      string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName: "Succès - Get calendar",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserget+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Test", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				return token, fmt.Sprintf("%d", calendarID), ""
			},
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessGetCalendar,
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Non authentifié",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserget+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Test", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				return "", fmt.Sprintf("%d", calendarID), ""
			},
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Erreur - Accès interdit",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserget+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Test", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				// Création d'un autre utilisateur
				otherEmail := fmt.Sprintf("othercalendaruser+%d@test.com", time.Now().UnixNano())
				otherUserID, otherResp := createUser(router, otherEmail, password, "Autre", "User")
				require.NotEqual(t, 0, otherUserID, "La création de l'autre utilisateur a échoué")
				require.Equal(t, http.StatusCreated, otherResp.StatusCode)
				otherToken, err := loginAndGetToken(router, otherEmail, password)
				require.NoError(t, err)
				return otherToken, fmt.Sprintf("%d", calendarID), ""
			},
			ExpectedStatus:  403,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Erreur - Calendar inexistant",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserget+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				return token, "99999", ""
			},
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrCalendarNotFound,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			testutils.ResetTestDB()
			router := setupTestRouter()
			token, calendarID, _ := testCase.Prepare(router)
			url := "/calendar/" + calendarID
			req := httptest.NewRequest("GET", url, nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()
			var jsonResp struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			body, _ := io.ReadAll(resp.Body)
			_ = json.Unmarshal(body, &jsonResp)
			require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
			require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
			if testCase.ExpectedMsg != "" {
				require.Contains(t, jsonResp.Message, testCase.ExpectedMsg)
			}
			if testCase.ExpectedError != "" {
				require.Contains(t, jsonResp.Error, testCase.ExpectedError)
			}
		})
	}
}

// --- Test table-driven pour PUT /calendar/:calendar_id ---
func TestUpdateCalendar(t *testing.T) {
	var TestCases = []struct {
		CaseName        string
		Prepare         func(router *gin.Engine) (token string, calendarID string, otherToken string)
		Payload         string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName: "Succès - Update calendar",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserupdate+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Update", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				return token, fmt.Sprintf("%d", calendarID), ""
			},
			Payload:         `{"title":"Calendrier Modifié","description":"Description modifiée"}`,
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessUpdateCalendar,
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Non authentifié",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserupdate+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Update", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				return "", fmt.Sprintf("%d", calendarID), ""
			},
			Payload:         `{"title":"Calendrier Modifié"}`,
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Erreur - Accès interdit",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserupdate+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Update", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				// Création d'un autre utilisateur
				otherEmail := fmt.Sprintf("othercalendaruserupdate+%d@test.com", time.Now().UnixNano())
				otherUserID, otherResp := createUser(router, otherEmail, password, "Autre", "User")
				require.NotEqual(t, 0, otherUserID, "La création de l'autre utilisateur a échoué")
				require.Equal(t, http.StatusCreated, otherResp.StatusCode)
				otherToken, err := loginAndGetToken(router, otherEmail, password)
				require.NoError(t, err)
				return otherToken, fmt.Sprintf("%d", calendarID), ""
			},
			Payload:         `{"title":"Calendrier Hacké"}`,
			ExpectedStatus:  403,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Erreur - Calendar inexistant",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserupdate+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				return token, "99999", ""
			},
			Payload:         `{"title":"Calendrier Modifié"}`,
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrCalendarNotFound,
		},
		{
			CaseName: "Erreur - Données invalides (titre manquant)",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserupdate+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Update", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				return token, fmt.Sprintf("%d", calendarID), ""
			},
			Payload:         `{"description":"Description sans titre"}`,
			ExpectedStatus:  400,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrInvalidData,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			testutils.ResetTestDB()
			router := setupTestRouter()
			token, calendarID, _ := testCase.Prepare(router)
			url := "/calendar/" + calendarID
			req := httptest.NewRequest("PUT", url, bytes.NewReader([]byte(testCase.Payload)))
			req.Header.Set("Content-Type", "application/json")
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()
			var jsonResp struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			body, _ := io.ReadAll(resp.Body)
			_ = json.Unmarshal(body, &jsonResp)
			require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
			require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
			if testCase.ExpectedMsg != "" {
				require.Contains(t, jsonResp.Message, testCase.ExpectedMsg)
			}
			if testCase.ExpectedError != "" {
				require.Contains(t, jsonResp.Error, testCase.ExpectedError)
			}
		})
	}
}

// --- Test table-driven pour DELETE /calendar/:calendar_id ---
func TestDeleteCalendar(t *testing.T) {
	var TestCases = []struct {
		CaseName        string
		Prepare         func(router *gin.Engine) (token string, calendarID string, otherToken string)
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName: "Succès - Delete calendar",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserdelete+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Delete", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				return token, fmt.Sprintf("%d", calendarID), ""
			},
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessDeleteCalendar,
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Non authentifié",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserdelete+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Delete", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				return "", fmt.Sprintf("%d", calendarID), ""
			},
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Erreur - Accès interdit",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserdelete+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				calendarID, resp := createCalendar(router, token, "Calendrier Delete", "Description")
				require.NotEqual(t, 0, calendarID, "La création du calendrier a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				// Création d'un autre utilisateur
				otherEmail := fmt.Sprintf("othercalendaruserdelete+%d@test.com", time.Now().UnixNano())
				otherUserID, otherResp := createUser(router, otherEmail, password, "Autre", "User")
				require.NotEqual(t, 0, otherUserID, "La création de l'autre utilisateur a échoué")
				require.Equal(t, http.StatusCreated, otherResp.StatusCode)
				otherToken, err := loginAndGetToken(router, otherEmail, password)
				require.NoError(t, err)
				return otherToken, fmt.Sprintf("%d", calendarID), ""
			},
			ExpectedStatus:  403,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Erreur - Calendar inexistant",
			Prepare: func(router *gin.Engine) (string, string, string) {
				email := fmt.Sprintf("calendaruserdelete+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				return token, "99999", ""
			},
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrCalendarNotFound,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			testutils.ResetTestDB()
			router := setupTestRouter()
			token, calendarID, _ := testCase.Prepare(router)
			url := "/calendar/" + calendarID
			req := httptest.NewRequest("DELETE", url, nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()
			var jsonResp struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			body, _ := io.ReadAll(resp.Body)
			_ = json.Unmarshal(body, &jsonResp)
			require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
			require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
			if testCase.ExpectedMsg != "" {
				require.Contains(t, jsonResp.Message, testCase.ExpectedMsg)
			}
			if testCase.ExpectedError != "" {
				require.Contains(t, jsonResp.Error, testCase.ExpectedError)
			}
		})
	}
}

// --- Test table-driven pour POST /calendar ---
func TestAddCalendar(t *testing.T) {
	var TestCases = []struct {
		CaseName        string
		Prepare         func(router *gin.Engine) (token string)
		Payload         string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName: "Succès - Création calendrier",
			Prepare: func(router *gin.Engine) string {
				email := fmt.Sprintf("calendaradd+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				return token
			},
			Payload:         `{"title":"Calendrier Test","description":"Description"}`,
			ExpectedStatus:  201,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessCreateCalendar,
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Non authentifié",
			Prepare: func(router *gin.Engine) string {
				return ""
			},
			Payload:         `{"title":"Calendrier Test","description":"Description"}`,
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Erreur - Données invalides (titre manquant)",
			Prepare: func(router *gin.Engine) string {
				email := fmt.Sprintf("calendaradd+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				userID, resp := createUser(router, email, password, "Jean", "Calendrier")
				require.NotEqual(t, 0, userID, "La création de l'utilisateur a échoué")
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				token, err := loginAndGetToken(router, email, password)
				require.NoError(t, err)
				return token
			},
			Payload:         `{"description":"Description sans titre"}`,
			ExpectedStatus:  400,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrInvalidData,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			testutils.ResetTestDB()
			router := setupTestRouter()
			token := testCase.Prepare(router)
			req := httptest.NewRequest("POST", "/calendar", bytes.NewReader([]byte(testCase.Payload)))
			req.Header.Set("Content-Type", "application/json")
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()
			var jsonResp struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			body, _ := io.ReadAll(resp.Body)
			_ = json.Unmarshal(body, &jsonResp)
			require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
			require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
			if testCase.ExpectedMsg != "" {
				require.Contains(t, jsonResp.Message, testCase.ExpectedMsg)
			}
			if testCase.ExpectedError != "" {
				require.Contains(t, jsonResp.Error, testCase.ExpectedError)
			}
		})
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	common.DB.Close()
	os.Exit(code)
}
