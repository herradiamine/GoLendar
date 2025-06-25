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

	// Configuration des routes pour les tests user_calendar avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })

	// Routes protégées par authentification
	router.POST("/calendar", middleware.AuthMiddleware(), func(c *gin.Context) { calendar.Calendar.Add(c) })
	router.GET("/calendar/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { calendar.Calendar.Get(c) })

	// Routes user_calendar avec authentification et accès au calendrier
	router.GET("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) { UserCalendar.Get(c) })
	router.GET("/user-calendar/:user_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { UserCalendar.GetByUser(c) })
	router.POST("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) { UserCalendar.Add(c) })
	router.PUT("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) { UserCalendar.Update(c) })
	router.DELETE("/user-calendar/:user_id/:calendar_id", middleware.AuthMiddleware(), middleware.UserExistsMiddleware("user_id"), middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) { UserCalendar.Delete(c) })

	return router
}

// --- Helpers mutualisés ---
func createUser(router http.Handler, email, password, firstname, lastname string) (int, string) {
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
	resp := w.Result()
	defer resp.Body.Close()
	var jsonResp struct {
		Success bool `json:"success"`
		Data    struct {
			UserID int `json:"user_id"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	return jsonResp.Data.UserID, jsonResp.Error
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

func createCalendar(router http.Handler, token, title, description string) (int, string) {
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
	resp := w.Result()
	defer resp.Body.Close()
	var jsonResp struct {
		Success bool `json:"success"`
		Data    struct {
			CalendarID int `json:"calendar_id"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	return jsonResp.Data.CalendarID, jsonResp.Error
}

// Helper pour créer un calendrier sans liaison user-calendar
func createOrphanCalendar(_ int, title, description string) int {
	res, err := common.DB.Exec("INSERT INTO calendar (title, description) VALUES (?, ?)", title, description)
	if err != nil {
		panic("Erreur création calendrier orphelin: " + err.Error())
	}
	id, _ := res.LastInsertId()
	return int(id)
}

// --- Table-driven tests pour toutes les routes user_calendar ---
func TestUserCalendar_AllRoutes_TableDriven(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()

	t.Run("POST /user-calendar/:user_id/:calendar_id (création)", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func(router http.Handler) (int, int, string, bool) // userID, calendarID, token, createBefore
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Création lien user-calendar",
				Setup: func(router http.Handler) (int, int, string, bool) {
					email := fmt.Sprintf("usercaladd+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID := createOrphanCalendar(userID, "Calendrier UC", "Desc") // SANS liaison
					return userID, calendarID, token, false                             // ne pas créer la liaison avant
				},
				ExpectedStatus:  201,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessCreateUserCalendar,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Conflit (déjà existant)",
				Setup: func(router http.Handler) (int, int, string, bool) {
					email := fmt.Sprintf("usercalconflict+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					return userID, calendarID, token, true // créer la liaison avant
				},
				ExpectedStatus:  409,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrUserCalendarConflict,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				userID, calendarID, token, createBefore := c.Setup(router)
				url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
				if createBefore {
					req := httptest.NewRequest("POST", url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
				}
				req := httptest.NewRequest("POST", url, nil)
				req.Header.Set("Authorization", "Bearer "+token)
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
				require.Equal(t, c.ExpectedStatus, resp.StatusCode)
				require.Equal(t, c.ExpectedSuccess, jsonResp.Success)
				if c.ExpectedMsg != "" {
					require.Contains(t, jsonResp.Message, c.ExpectedMsg)
				}
				if c.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, c.ExpectedError)
				}
			})
		}
	})

	// --- Table-driven pour GET /user-calendar/:user_id/:calendar_id ---
	t.Run("GET /user-calendar/:user_id/:calendar_id", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func(router http.Handler) (int, int, string, bool) // userID, calendarID, token, createLink
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Get user-calendar link",
				Setup: func(router http.Handler) (int, int, string, bool) {
					email := fmt.Sprintf("usercalget+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					return userID, calendarID, token, true // créer la liaison
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Non authentifié",
				Setup: func(router http.Handler) (int, int, string, bool) {
					email := fmt.Sprintf("usercalgetunauth+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					return userID, calendarID, "", true // liaison créée mais pas de token
				},
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName: "Erreur - Liaison inexistante",
				Setup: func(router http.Handler) (int, int, string, bool) {
					email := fmt.Sprintf("usercalgetnotfound+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID := createOrphanCalendar(userID, "Calendrier UC", "Desc") // SANS liaison
					return userID, calendarID, token, false                             // ne pas créer la liaison
				},
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserCalendarNotFound,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				userID, calendarID, token, createLink := c.Setup(router)
				if createLink {
					url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
					req := httptest.NewRequest("POST", url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
				}
				url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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
					Error   string `json:"error"`
				}
				body, _ := io.ReadAll(resp.Body)
				_ = json.Unmarshal(body, &jsonResp)
				require.Equal(t, c.ExpectedStatus, resp.StatusCode)
				require.Equal(t, c.ExpectedSuccess, jsonResp.Success)
				if c.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, c.ExpectedError)
				}
			})
		}
	})

	// --- Table-driven pour GET /user-calendar/:user_id (listing) ---
	t.Run("GET /user-calendar/:user_id", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func(router http.Handler) (int, string) // userID, token
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Listing user-calendar",
				Setup: func(router http.Handler) (int, string) {
					email := fmt.Sprintf("usercallist+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					// Créer la liaison
					url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
					req := httptest.NewRequest("POST", url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					return userID, token
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Non authentifié",
				Setup: func(router http.Handler) (int, string) {
					email := fmt.Sprintf("usercallistunauth+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					return userID, ""
				},
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				userID, token := c.Setup(router)
				url := fmt.Sprintf("/user-calendar/%d", userID)
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
					Error   string `json:"error"`
				}
				body, _ := io.ReadAll(resp.Body)
				_ = json.Unmarshal(body, &jsonResp)
				require.Equal(t, c.ExpectedStatus, resp.StatusCode)
				require.Equal(t, c.ExpectedSuccess, jsonResp.Success)
				if c.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, c.ExpectedError)
				}
			})
		}
	})

	// --- Table-driven pour PUT /user-calendar/:user_id/:calendar_id (update) ---
	t.Run("PUT /user-calendar/:user_id/:calendar_id", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func(router http.Handler) (int, int, string) // userID, calendarID, token
			Payload         string
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Update permissions",
				Setup: func(router http.Handler) (int, int, string) {
					email := fmt.Sprintf("usercalupd+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					// Créer la liaison
					url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
					req := httptest.NewRequest("POST", url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					return userID, calendarID, token
				},
				Payload:         `{"permissions":"read_write"}`,
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessUpdateUserCalendar,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Non authentifié",
				Setup: func(router http.Handler) (int, int, string) {
					email := fmt.Sprintf("usercalupdauth+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
					req := httptest.NewRequest("POST", url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					return userID, calendarID, ""
				},
				Payload:         `{"permissions":"read_write"}`,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				userID, calendarID, token := c.Setup(router)
				url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
				req := httptest.NewRequest("PUT", url, bytes.NewReader([]byte(c.Payload)))
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
				require.Equal(t, c.ExpectedStatus, resp.StatusCode)
				require.Equal(t, c.ExpectedSuccess, jsonResp.Success)
				if c.ExpectedMsg != "" {
					require.Contains(t, jsonResp.Message, c.ExpectedMsg)
				}
				if c.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, c.ExpectedError)
				}
			})
		}
	})

	// --- Table-driven pour DELETE /user-calendar/:user_id/:calendar_id ---
	t.Run("DELETE /user-calendar/:user_id/:calendar_id", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func(router http.Handler) (int, int, string) // userID, calendarID, token
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Suppression user-calendar",
				Setup: func(router http.Handler) (int, int, string) {
					email := fmt.Sprintf("usercaldel+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					// Créer la liaison
					url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
					req := httptest.NewRequest("POST", url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					return userID, calendarID, token
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessDeleteUserCalendar,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Non authentifié",
				Setup: func(router http.Handler) (int, int, string) {
					email := fmt.Sprintf("usercaldelunauth+%d@test.com", time.Now().UnixNano())
					userID, _ := createUser(router, email, "motdepasse123", "Jean", "UC")
					token, _ := loginAndGetToken(router, email, "motdepasse123")
					calendarID, _ := createCalendar(router, token, "Calendrier UC", "Desc")
					url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
					req := httptest.NewRequest("POST", url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					return userID, calendarID, ""
				},
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				userID, calendarID, token := c.Setup(router)
				url := fmt.Sprintf("/user-calendar/%d/%d", userID, calendarID)
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
				require.Equal(t, c.ExpectedStatus, resp.StatusCode)
				require.Equal(t, c.ExpectedSuccess, jsonResp.Success)
				if c.ExpectedMsg != "" {
					require.Contains(t, jsonResp.Message, c.ExpectedMsg)
				}
				if c.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, c.ExpectedError)
				}
			})
		}
	})
}

func TestMain(m *testing.M) {
	code := m.Run()
	common.DB.Close()
	os.Exit(code)
}
