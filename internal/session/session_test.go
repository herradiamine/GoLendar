package session

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

func TestSessionAuthentication(t *testing.T) {
	router := setupTestRouter()
	var userToken string
	var refreshToken string
	uniqueEmail := fmt.Sprintf("session.user+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Session",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	t.Run("Login User", func(t *testing.T) {
		payload := common.LoginRequest{
			Email:    uniqueEmail,
			Password: "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessLogin, response.Message)

		// Récupérer les tokens
		if data, ok := response.Data.(map[string]interface{}); ok {
			if token, ok := data["session_token"]; ok {
				userToken = token.(string)
			}
			if refresh, ok := data["refresh_token"]; ok {
				refreshToken = refresh.(string)
			}
		}
		require.NotEmpty(t, userToken)
		require.NotEmpty(t, refreshToken)
	})

	t.Run("Get User Info", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Get User Sessions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Refresh Token", func(t *testing.T) {
		payload := RefreshTokenRequest{
			RefreshToken: refreshToken,
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessRefreshToken, response.Message)

		// Récupérer le nouveau token
		if data, ok := response.Data.(map[string]interface{}); ok {
			if token, ok := data["session_token"]; ok {
				userToken = token.(string)
			}
		}
	})

	t.Run("Logout User", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessLogout, response.Message)
	})
}

func TestSessionErrorCases(t *testing.T) {
	router := setupTestRouter()
	uniqueEmail := fmt.Sprintf("error.session+%d@test.com", time.Now().UnixNano())

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
	}

	t.Run("Login with Invalid Credentials", func(t *testing.T) {
		payload := common.LoginRequest{
			Email:    uniqueEmail,
			Password: "wrongpassword",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrInvalidCredentials, response.Error)
	})

	t.Run("Refresh Token with Invalid Token", func(t *testing.T) {
		payload := RefreshTokenRequest{
			RefreshToken: "invalid-refresh-token",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
	})

	t.Run("Refresh Token with Empty Token", func(t *testing.T) {
		payload := RefreshTokenRequest{
			RefreshToken: "",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
	})

	t.Run("Logout without Token", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/logout", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, 401, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrSessionInvalid, response.Error)
	})

	t.Run("Get Sessions without Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/sessions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
	})

	t.Run("Delete Session without Token", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/auth/sessions/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
	})
}

func TestSessionSecurity(t *testing.T) {
	router := setupTestRouter()
	var userToken string
	uniqueEmail := fmt.Sprintf("security.session+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests de sécurité
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Security",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Login pour obtenir le token
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

	t.Run("Access Protected Route with Valid Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Access Protected Route with Invalid Token Format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/sessions", nil)
		req.Header.Set("Authorization", "InvalidToken")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
	})

	t.Run("Access Protected Route with Empty Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/sessions", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
	})
}

func TestDeleteSession_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/auth/sessions/:session_id", func(c *gin.Context) {
		Session.DeleteSession(c)
	})
	req, _ := http.NewRequest("DELETE", "/auth/sessions/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusInternalServerError {
		t.Errorf("DeleteSession sans user: code HTTP = %d, want 401 ou 500", w.Code)
	}
}

func TestDeleteSession_MissingSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth_user", common.User{UserID: 1})
	})
	r.DELETE("/auth/sessions/", func(c *gin.Context) {
		Session.DeleteSession(c)
	})
	req, _ := http.NewRequest("DELETE", "/auth/sessions/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("DeleteSession sans session_id: code HTTP = %d, want 400", w.Code)
	}
}

func TestValidateSession_TokenInconnu(t *testing.T) {
	_, err := Session.ValidateSession("token-inconnu")
	if err == nil {
		t.Error("ValidateSession devrait retourner une erreur pour un token inconnu")
	}
}

func TestValidateSession_DBError(t *testing.T) {
	if common.DB != nil {
		_ = common.DB.Close()
	}
	_, err := Session.ValidateSession("token-test")
	if err == nil {
		t.Error("ValidateSession devrait retourner une erreur si la DB est fermée")
	}
}
