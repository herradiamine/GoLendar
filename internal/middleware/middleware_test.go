package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/common"
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

	// Configuration des routes pour tester les middlewares
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })

	// Routes pour tester AuthMiddleware
	router.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Protected route accessed"})
	})

	// Routes pour tester AdminMiddleware
	router.GET("/admin", AuthMiddleware(), AdminMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Admin route accessed"})
	})

	// Routes pour tester RoleMiddleware
	router.GET("/moderator", AuthMiddleware(), RoleMiddleware("moderator"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Moderator route accessed"})
	})

	// Routes pour tester UserExistsMiddleware
	router.GET("/user/:user_id", AuthMiddleware(), UserExistsMiddleware("user_id"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "User exists"})
	})

	// Routes pour tester CalendarExistsMiddleware
	router.GET("/calendar/:calendar_id", AuthMiddleware(), CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Calendar exists"})
	})

	// Routes pour tester EventExistsMiddleware
	router.GET("/event/:event_id", AuthMiddleware(), EventExistsMiddleware("event_id"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Event exists"})
	})

	// Routes pour tester RoleExistsMiddleware
	router.GET("/role/:role_id", AuthMiddleware(), RoleExistsMiddleware("role_id"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Role exists"})
	})

	return router
}

func TestAuthMiddleware(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	var userToken string
	uniqueEmail := fmt.Sprintf("auth.middleware+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Auth",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		token, err := loginAndGetToken(router, uniqueEmail, "motdepasse123")
		require.NoError(t, err)
		require.NotEmpty(t, token)
		userToken = token
	}

	t.Run("Access Protected Route with Valid Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, "Protected route accessed", response["message"])
	})

	t.Run("Access Protected Route without Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotAuthenticated, response.Error)
	})

	t.Run("Access Protected Route with Invalid Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrSessionInvalid, response.Error)
	})

	t.Run("Access Protected Route with Malformed Authorization Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrSessionInvalid, response.Error)
	})
}

func TestAdminMiddleware(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	var adminToken, userToken string
	adminEmail := fmt.Sprintf("admin.middleware+%d@test.com", time.Now().UnixNano())
	userEmail := fmt.Sprintf("user.middleware+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur admin
	{
		payload := common.CreateUserRequest{
			Lastname:  "Admin",
			Firstname: "Middleware",
			Email:     adminEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Récupérer l'ID de l'utilisateur admin
		var adminUserID int
		row := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", adminEmail)
		err := row.Scan(&adminUserID)
		require.NoError(t, err)

		// Créer le rôle admin s'il n'existe pas
		var adminRoleID int
		row = common.DB.QueryRow("SELECT role_id FROM roles WHERE name = 'admin'")
		err = row.Scan(&adminRoleID)
		if err != nil {
			// Le rôle admin n'existe pas, on le crée
			res, err2 := common.DB.Exec("INSERT INTO roles (name, description) VALUES ('admin', 'Administrateur')")
			require.NoError(t, err2)
			id, _ := res.LastInsertId()
			adminRoleID = int(id)
		}

		// Attribuer le rôle admin à l'utilisateur
		_, err = common.DB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", adminUserID, adminRoleID)
		if err != nil {
			// Si l'insertion échoue, on vérifie si la liaison existe déjà
			var count int
			row = common.DB.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = ? AND role_id = ?", adminUserID, adminRoleID)
			err2 := row.Scan(&count)
			if err2 != nil || count == 0 {
				require.NoError(t, err)
			}
		}

		token, err := loginAndGetToken(router, adminEmail, "motdepasse123")
		require.NoError(t, err)
		require.NotEmpty(t, token)
		adminToken = token
	}

	// Créer un utilisateur normal
	{
		payload := common.CreateUserRequest{
			Lastname:  "User",
			Firstname: "Middleware",
			Email:     userEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		token, err := loginAndGetToken(router, userEmail, "motdepasse123")
		require.NoError(t, err)
		require.NotEmpty(t, token)
		userToken = token
	}

	t.Run("Access Admin Route with Admin User", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Access Admin Route with Regular User", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
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

	t.Run("Access Admin Route without Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotAuthenticated, response.Error)
	})
}

func TestRoleMiddleware(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	var userToken string
	uniqueEmail := fmt.Sprintf("role.middleware+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Role",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		token, err := loginAndGetToken(router, uniqueEmail, "motdepasse123")
		require.NoError(t, err)
		require.NotEmpty(t, token)
		userToken = token
	}

	t.Run("Access Role-Specific Route without Required Role", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/moderator", nil)
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

func TestExistsMiddlewares(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	var userToken string
	uniqueEmail := fmt.Sprintf("exists.middleware+%d@test.com", time.Now().UnixNano())

	// Créer un utilisateur pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Test",
			Firstname: "Exists",
			Email:     uniqueEmail,
			Password:  "motdepasse123",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		token, err := loginAndGetToken(router, uniqueEmail, "motdepasse123")
		require.NoError(t, err)
		require.NotEmpty(t, token)
		userToken = token
	}

	t.Run("UserExistsMiddleware with Non-existent User", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/999", nil)
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

	t.Run("CalendarExistsMiddleware with Non-existent Calendar", func(t *testing.T) {
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

	t.Run("EventExistsMiddleware with Non-existent Event", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/event/999", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrEventNotFound, response.Error)
	})

	t.Run("RoleExistsMiddleware with Non-existent Role", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/role/999", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrRoleNotFound, response.Error)
	})

	t.Run("UserExistsMiddleware with Invalid User ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/invalid", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, "ID utilisateur invalide", response.Error)
	})
}

func TestExtractTokenFromHeader(t *testing.T) {
	token := "Bearer abcdef123456"
	if res := extractTokenFromHeader(token); res != "abcdef123456" {
		t.Errorf("extractTokenFromHeader(%q) = %q, want %q", token, res, "abcdef123456")
	}
	if res := extractTokenFromHeader(""); res != "" {
		t.Errorf("extractTokenFromHeader(\"\") = %q, want empty string", res)
	}
	if res := extractTokenFromHeader("Bearer"); res != "" {
		t.Errorf("extractTokenFromHeader('Bearer') = %q, want empty string", res)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggingMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("LoggingMiddleware: code HTTP = %d, want 200", w.Code)
	}
}

func TestOptionalAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OptionalAuthMiddleware())
	r.GET("/public", func(c *gin.Context) {
		c.String(200, "ok")
	})
	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("OptionalAuthMiddleware sans token: code HTTP = %d, want 200", w.Code)
	}
}

func TestUserCanAccessCalendarMiddleware_NoUserOrCalendar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(UserCanAccessCalendarMiddleware())
	r.GET("/calendar", func(c *gin.Context) {
		c.String(200, "ok")
	})
	// Pas d'utilisateur ni de calendrier dans le contexte
	req, _ := http.NewRequest("GET", "/calendar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// On attend une absence de réponse 200
	if w.Code == 200 {
		t.Error("UserCanAccessCalendarMiddleware devrait bloquer sans user/calendar dans le contexte")
	}
}

func TestRolesMiddleware_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// Simule un utilisateur dans le contexte
		c.Set("auth_user", common.User{UserID: 1})
	})
	r.Use(RolesMiddleware("admin", "moderator"))
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.False(t, response.Success)
	require.Equal(t, common.ErrInsufficientPermissions, response.Error)
}

func loginAndGetToken(router *gin.Engine, email string, password string) (string, error) {
	payload := common.LoginRequest{
		Email:    email,
		Password: password,
	}
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		return "", err
	}
	if !response.Success {
		return "", fmt.Errorf("login failed: %v", response.Error)
	}
	if data, ok := response.Data.(map[string]interface{}); ok {
		if token, ok := data["session_token"]; ok {
			return token.(string), nil
		}
	}
	return "", fmt.Errorf("token not found in response")
}
