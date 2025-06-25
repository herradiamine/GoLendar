package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/internal/session"
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

	// Configuration des routes pour les tests utilisateur avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })
	router.POST("/auth/refresh", func(c *gin.Context) { session.Session.RefreshToken(c) })

	// Routes protégées par authentification
	router.GET("/user/me", middleware.AuthMiddleware(), func(c *gin.Context) { User.Get(c) })
	router.PUT("/user/me", middleware.AuthMiddleware(), func(c *gin.Context) { User.Update(c) })
	router.DELETE("/user/me", middleware.AuthMiddleware(), func(c *gin.Context) { User.Delete(c) })
	router.POST("/auth/logout", middleware.AuthMiddleware(), func(c *gin.Context) { session.Session.Logout(c) })
	router.GET("/auth/me", middleware.AuthMiddleware(), func(c *gin.Context) { User.GetUserWithRoles(c) })
	router.GET("/auth/sessions", middleware.AuthMiddleware(), func(c *gin.Context) { session.Session.GetUserSessions(c) })

	// Routes admin
	router.GET("/user/:user_id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.Get(c) })
	router.PUT("/user/:user_id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.Update(c) })
	router.DELETE("/user/:user_id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.Delete(c) })
	router.GET("/user/:user_id/with-roles", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.GetUserWithRoles(c) })

	return router
}

func TestUserCRUD(t *testing.T) {
	router := setupTestRouter()
	var userToken string
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

		require.Equal(t, http.StatusCreated, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessCreateUser, response.Message)
	})

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

		// Récupérer le token
		if data, ok := response.Data.(map[string]interface{}); ok {
			if token, ok := data["session_token"]; ok {
				userToken = token.(string)
			}
		}
		require.NotEmpty(t, userToken)
	})

	t.Run("Get User Me", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/me", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Get Auth Me", func(t *testing.T) {
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

	t.Run("Update User Me", func(t *testing.T) {
		payload := common.UpdateUserRequest{
			Lastname:  common.StringPtr("Martin"),
			Firstname: common.StringPtr("Pierre"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessUpdateUser, response.Message)
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

func TestUserErrorCases(t *testing.T) {
	router := setupTestRouter()
	var userToken string
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

	t.Run("Get User Me Without Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/me", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotAuthenticated, response.Error)
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

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Contains(t, response.Error, common.ErrInvalidData)
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

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Contains(t, response.Error, common.ErrInvalidData)
	})

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

	t.Run("Access Admin Route Without Admin Role", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/999", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrInsufficientPermissions, response.Error)
	})
}
