package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/common"
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

// Structures manquantes pour les tests
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Fonction d'authentification simple pour les tests
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrSessionInvalid,
			})
			c.Abort()
			return
		}
		// Vérification stricte du token pour les tests
		if len(token) < 7 || token[:7] != "Bearer " || len(token[7:]) == 0 {
			c.JSON(http.StatusUnauthorized, common.JSONResponse{
				Success: false,
				Error:   common.ErrSessionInvalid,
			})
			c.Abort()
			return
		}
		// Pour les tests, on crée un utilisateur factice
		user := common.User{
			UserID:    1,
			Lastname:  "Test",
			Firstname: "User",
			Email:     "test@test.com",
			CreatedAt: time.Now(),
		}
		c.Set("user", user)
		c.Next()
	}
}

// --- Helpers mutualisés ---
func createUser(router http.Handler, email, password, firstname, lastname string) *http.Response {
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
	return w.Result()
}

func loginAndGetTokens(router http.Handler, email, password string) (string, string, error) {
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
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			SessionToken string `json:"session_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	if !jsonResp.Success || jsonResp.Data.SessionToken == "" {
		return "", "", fmt.Errorf("login failed: %s", jsonResp.Error)
	}
	return jsonResp.Data.SessionToken, jsonResp.Data.RefreshToken, nil
}

func setupTestRouter() *gin.Engine {
	router := testutils.SetupTestRouter()

	// Configuration des routes pour les tests sessions avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { Session.Login(c) })
	router.POST("/auth/refresh", func(c *gin.Context) { Session.RefreshToken(c) })

	// Routes protégées par authentification
	router.POST("/auth/logout", authMiddleware(), func(c *gin.Context) { Session.Logout(c) })
	router.GET("/auth/me", authMiddleware(), func(c *gin.Context) { Session.GetUserSessions(c) })
	router.GET("/auth/sessions", authMiddleware(), func(c *gin.Context) { Session.GetUserSessions(c) })
	router.DELETE("/auth/sessions/:session_id", authMiddleware(), func(c *gin.Context) { Session.DeleteSession(c) })

	return router
}

// --- Table-driven tests pour toutes les routes session ---
func TestSession_AllRoutes_TableDriven(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()

	// Création d'un utilisateur de test
	email := fmt.Sprintf("sessionuser+%d@test.com", time.Now().UnixNano())
	password := "motdepasse123"
	_ = createUser(router, email, password, "Jean", "Session")
	token, _, err := loginAndGetTokens(router, email, password)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// --- Table-driven pour /auth/login ---
	t.Run("POST /auth/login", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func() (string, string) // retourne (email, password)
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Login",
				Setup: func() (string, string) {
					router := setupTestRouter()
					email := fmt.Sprintf("sessionuser+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_ = createUser(router, email, password, "Jean", "Session")
					return email, password
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessLogin,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Credentials invalides",
				Setup: func() (string, string) {
					return "invalid@test.com", "wrongpassword"
				},
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrInvalidCredentials,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				email, password := c.Setup()
				router := setupTestRouter()
				payload := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
				req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte(payload)))
				req.Header.Set("Content-Type", "application/json")
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

	// --- Table-driven pour /auth/refresh ---
	t.Run("POST /auth/refresh", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func() (string, string) // retourne (refreshToken, password)
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Refresh token",
				Setup: func() (string, string) {
					router := setupTestRouter()
					email := fmt.Sprintf("sessionrefresh+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_ = createUser(router, email, password, "Jean", "Session")
					_, refreshToken, _ := loginAndGetTokens(router, email, password)
					return refreshToken, password
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessRefreshToken,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Token refresh invalide",
				Setup: func() (string, string) {
					return "invalid-refresh-token", "motdepasse123"
				},
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrSessionInvalid,
			},
			{
				CaseName: "Erreur - Token refresh vide",
				Setup: func() (string, string) {
					return "", "motdepasse123"
				},
				ExpectedStatus:  400,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrInvalidData,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				refreshToken, _ := c.Setup()
				router := setupTestRouter()
				payload := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)
				req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader([]byte(payload)))
				req.Header.Set("Content-Type", "application/json")
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

	// --- Table-driven pour /auth/logout ---
	t.Run("POST /auth/logout", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func() string // retourne le token
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
		}{
			{
				CaseName: "Succès - Logout",
				Setup: func() string {
					router := setupTestRouter()
					email := fmt.Sprintf("sessionlogout+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_ = createUser(router, email, password, "Jean", "Session")
					_, _, _ = loginAndGetTokens(router, email, password)
					return ""
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessLogout,
				ExpectedError:   "",
			},
			{
				CaseName: "Erreur - Non authentifié",
				Setup: func() string {
					return ""
				},
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrSessionInvalid,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				token := c.Setup()
				router := setupTestRouter()
				req := httptest.NewRequest("POST", "/auth/logout", nil)
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

	// --- Table-driven pour /auth/me et /auth/sessions ---
	t.Run("GET /auth/me et /auth/sessions", func(t *testing.T) {
		endpoints := []string{"/auth/me", "/auth/sessions"}
		for _, endpoint := range endpoints {
			cases := []struct {
				CaseName        string
				Setup           func() string // retourne le token
				ExpectedStatus  int
				ExpectedSuccess bool
				ExpectedError   string
			}{
				{
					CaseName: "Succès - Accès protégé",
					Setup: func() string {
						router := setupTestRouter()
						email := fmt.Sprintf("sessionme+%d@test.com", time.Now().UnixNano())
						password := "motdepasse123"
						_ = createUser(router, email, password, "Jean", "Session")
						_, _, _ = loginAndGetTokens(router, email, password)
						return ""
					},
					ExpectedStatus:  200,
					ExpectedSuccess: true,
					ExpectedError:   "",
				},
				{
					CaseName: "Erreur - Non authentifié",
					Setup: func() string {
						return ""
					},
					ExpectedStatus:  401,
					ExpectedSuccess: false,
					ExpectedError:   common.ErrSessionInvalid,
				},
			}
			for _, c := range cases {
				t.Run(endpoint+"/"+c.CaseName, func(t *testing.T) {
					token := c.Setup()
					router := setupTestRouter()
					req := httptest.NewRequest("GET", endpoint, nil)
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
		}
	})

	// --- Table-driven pour DELETE /auth/sessions/:session_id ---
	t.Run("DELETE /auth/sessions/:session_id", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func() (string, string) // retourne (token, sessionID)
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName: "Erreur - Non authentifié",
				Setup: func() (string, string) {
					return "", "1"
				},
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrSessionInvalid,
			},
			{
				CaseName: "Erreur - Session inexistante",
				Setup: func() (string, string) {
					router := setupTestRouter()
					email := fmt.Sprintf("sessiondelsess+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_ = createUser(router, email, password, "Jean", "Session")
					_, _, _ = loginAndGetTokens(router, email, password)
					return "", "99999"
				},
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrSessionNotFound,
			},
		}
		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				token, sessionID := c.Setup()
				router := setupTestRouter()
				url := "/auth/sessions/" + sessionID
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
}

func TestMain(m *testing.M) {
	testutils.SetupTestDB()
	code := m.Run()
	common.DB.Close()
	os.Exit(code)
}
