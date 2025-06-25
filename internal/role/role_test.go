package role

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

	// Configuration des routes pour les tests rôles avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })

	// Routes admin pour les rôles
	router.GET("/roles", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) { Role.ListRoles(c) })
	router.GET("/roles/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.RoleExistsMiddleware("id"), func(c *gin.Context) { Role.GetRole(c) })
	router.POST("/roles", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) { Role.CreateRole(c) })
	router.PUT("/roles/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.RoleExistsMiddleware("id"), func(c *gin.Context) { Role.UpdateRole(c) })
	router.DELETE("/roles/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.RoleExistsMiddleware("id"), func(c *gin.Context) { Role.DeleteRole(c) })
	router.POST("/roles/assign", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) { Role.AssignRole(c) })
	router.POST("/roles/revoke", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) { Role.RevokeRole(c) })
	router.GET("/roles/user/:user_id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { Role.GetUserRoles(c) })

	return router
}

func TestRoleCRUD(t *testing.T) {
	router := setupTestRouter()
	var roleID int
	var adminToken string
	adminEmail := fmt.Sprintf("admin.role+%d@test.com", time.Now().UnixNano())
	timestamp := time.Now().UnixNano()

	// Créer un utilisateur admin pour les tests
	{
		payload := common.CreateUserRequest{
			Lastname:  "Admin",
			Firstname: "Role",
			Email:     adminEmail,
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
				userID := int(id.(float64))
				// Vérifier si le rôle admin existe déjà
				var roleIDFromDB int
				err := common.DB.QueryRow("SELECT role_id FROM roles WHERE name = 'admin' AND deleted_at IS NULL").Scan(&roleIDFromDB)
				if err != nil {
					// Créer le rôle admin s'il n'existe pas
					result, err := common.DB.Exec("INSERT INTO roles (name, description, created_at) VALUES (?, ?, NOW())", "admin", "Rôle administrateur")
					if err == nil {
						roleIDFromDB64, _ := result.LastInsertId()
						roleIDFromDB = int(roleIDFromDB64)
					}
				}
				if roleIDFromDB > 0 {
					// Attribuer le rôle admin à l'utilisateur
					_, _ = common.DB.Exec("INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE updated_at = NOW()", userID, roleIDFromDB)
				}
			}
		}
	}

	// Login pour obtenir un token
	{
		payload := common.LoginRequest{
			Email:    adminEmail,
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
				adminToken = token.(string)
			}
		}
	}

	t.Run("Create Role", func(t *testing.T) {
		roleName := fmt.Sprintf("moderator_%d", timestamp)
		payload := common.CreateRoleRequest{
			Name:        roleName,
			Description: common.StringPtr("Rôle modérateur"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessCreateRole, response.Message)

		// Récupérer l'ID du rôle créé pour les tests suivants
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}
	})

	t.Run("Get Role", func(t *testing.T) {
		url := fmt.Sprintf("/roles/%d", roleID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("List Roles", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/roles", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Update Role", func(t *testing.T) {
		newRoleName := fmt.Sprintf("super_admin_%d", timestamp)
		payload := common.UpdateRoleRequest{
			Name:        common.StringPtr(newRoleName),
			Description: common.StringPtr("Rôle super administrateur"),
		}
		jsonData, _ := json.Marshal(payload)
		url := fmt.Sprintf("/roles/%d", roleID)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessUpdateRole, response.Message)
	})

	t.Run("Delete Role", func(t *testing.T) {
		url := fmt.Sprintf("/roles/%d", roleID)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessDeleteRole, response.Message)
	})
}

func TestRoleAssignment(t *testing.T) {
	router := setupTestRouter()
	var roleID int
	var userID int
	var adminToken string
	adminEmail := fmt.Sprintf("admin.assign+%d@test.com", time.Now().UnixNano())
	userEmail := fmt.Sprintf("user.assign+%d@test.com", time.Now().UnixNano())
	timestamp := time.Now().UnixNano()

	// Créer un utilisateur admin
	{
		payload := common.CreateUserRequest{
			Lastname:  "Admin",
			Firstname: "Assign",
			Email:     adminEmail,
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
				adminUserID := int(id.(float64))
				// Vérifier si le rôle admin existe déjà
				var roleIDFromDB int
				err := common.DB.QueryRow("SELECT role_id FROM roles WHERE name = 'admin' AND deleted_at IS NULL").Scan(&roleIDFromDB)
				if err != nil {
					// Créer le rôle admin s'il n'existe pas
					result, err := common.DB.Exec("INSERT INTO roles (name, description, created_at) VALUES (?, ?, NOW())", "admin", "Rôle administrateur")
					if err == nil {
						roleIDFromDB64, _ := result.LastInsertId()
						roleIDFromDB = int(roleIDFromDB64)
					}
				}
				if roleIDFromDB > 0 {
					// Attribuer le rôle admin à l'utilisateur admin
					_, _ = common.DB.Exec("INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE updated_at = NOW()", adminUserID, roleIDFromDB)
				}
			}
		}
	}

	// Créer un utilisateur normal
	{
		payload := common.CreateUserRequest{
			Lastname:  "User",
			Firstname: "Assign",
			Email:     userEmail,
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

	// Login admin
	{
		payload := common.LoginRequest{
			Email:    adminEmail,
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
				adminToken = token.(string)
			}
		}
	}

	// Créer un rôle
	{
		roleName := fmt.Sprintf("editor_%d", timestamp)
		payload := common.CreateRoleRequest{
			Name:        roleName,
			Description: common.StringPtr("Rôle éditeur"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}
	}

	t.Run("Assign Role to User", func(t *testing.T) {
		payload := common.AssignRoleRequest{
			UserID: userID,
			RoleID: roleID,
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles/assign", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessAssignRole, response.Message)
	})

	t.Run("Get User Roles", func(t *testing.T) {
		url := fmt.Sprintf("/roles/user/%d", userID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
	})

	t.Run("Revoke Role from User", func(t *testing.T) {
		payload := common.RevokeRoleRequest{
			UserID: userID,
			RoleID: roleID,
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles/revoke", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.True(t, response.Success)
		require.Equal(t, common.MsgSuccessRevokeRole, response.Message)
	})
}

func TestRoleErrorCases(t *testing.T) {
	router := setupTestRouter()
	var roleID int
	var userID int
	var adminToken string
	var userToken string
	adminEmail := fmt.Sprintf("admin.error+%d@test.com", time.Now().UnixNano())
	userEmail := fmt.Sprintf("user.error+%d@test.com", time.Now().UnixNano())
	timestamp := time.Now().UnixNano()

	// Créer un utilisateur admin
	{
		payload := common.CreateUserRequest{
			Lastname:  "Admin",
			Firstname: "Error",
			Email:     adminEmail,
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
				adminUserID := int(id.(float64))
				// Vérifier si le rôle admin existe déjà
				var roleIDFromDB int
				err := common.DB.QueryRow("SELECT role_id FROM roles WHERE name = 'admin' AND deleted_at IS NULL").Scan(&roleIDFromDB)
				if err != nil {
					// Créer le rôle admin s'il n'existe pas
					result, err := common.DB.Exec("INSERT INTO roles (name, description, created_at) VALUES (?, ?, NOW())", "admin", "Rôle administrateur")
					if err == nil {
						roleIDFromDB64, _ := result.LastInsertId()
						roleIDFromDB = int(roleIDFromDB64)
					}
				}
				if roleIDFromDB > 0 {
					// Attribuer le rôle admin à l'utilisateur admin
					_, _ = common.DB.Exec("INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE updated_at = NOW()", adminUserID, roleIDFromDB)
				}
			}
		}
	}

	// Créer un utilisateur normal
	{
		payload := common.CreateUserRequest{
			Lastname:  "User",
			Firstname: "Error",
			Email:     userEmail,
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

	// Login admin
	{
		payload := common.LoginRequest{
			Email:    adminEmail,
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
				adminToken = token.(string)
			}
		}
	}

	// Login user
	{
		payload := common.LoginRequest{
			Email:    userEmail,
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

	// Créer un rôle pour les tests
	{
		roleName := fmt.Sprintf("test_role_%d", timestamp)
		payload := common.CreateRoleRequest{
			Name:        roleName,
			Description: common.StringPtr("Rôle de test"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var response common.JSONResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		if data, ok := response.Data.(map[string]interface{}); ok {
			if id, ok := data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}
	}

	t.Run("Create Role Without Token", func(t *testing.T) {
		payload := common.CreateRoleRequest{
			Name:        "unauthorized_role",
			Description: common.StringPtr("Rôle non autorisé"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
	})

	t.Run("Create Role Without Admin Rights", func(t *testing.T) {
		payload := common.CreateRoleRequest{
			Name:        "user_role",
			Description: common.StringPtr("Rôle utilisateur"),
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
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

	t.Run("Get Non-existent Role", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/roles/999", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrRoleNotFound, response.Error)
	})

	t.Run("Create Role with Invalid Data", func(t *testing.T) {
		payload := map[string]interface{}{
			// Name manquant
			"description": "Description sans nom",
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Contains(t, response.Error, common.ErrInvalidData)
	})

	t.Run("Assign Role with Non-existent User", func(t *testing.T) {
		payload := common.AssignRoleRequest{
			UserID: 999,
			RoleID: roleID,
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles/assign", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrUserNotFound, response.Error)
	})

	t.Run("Assign Non-existent Role", func(t *testing.T) {
		payload := common.AssignRoleRequest{
			UserID: userID,
			RoleID: 999,
		}
		jsonData, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/roles/assign", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response common.JSONResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.False(t, response.Success)
		require.Equal(t, common.ErrRoleNotFound, response.Error)
	})
}
