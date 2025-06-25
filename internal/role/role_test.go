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

// --- Helpers mutualisés pour la refonte table-driven ---
func createAdmin(router http.Handler, email, password, firstname, lastname string) (string, error) {
	// Crée l'utilisateur
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

	// Vérifie que l'utilisateur a été créé
	var userID int
	row := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", email)
	err := row.Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %v", err)
	}

	// Crée le rôle admin s'il n'existe pas
	var adminRoleID int
	row = common.DB.QueryRow("SELECT role_id FROM roles WHERE name = 'admin'")
	err = row.Scan(&adminRoleID)
	if err != nil {
		// Le rôle admin n'existe pas, on le crée
		res, err2 := common.DB.Exec("INSERT INTO roles (name, description) VALUES ('admin', 'Administrateur')")
		if err2 != nil {
			return "", fmt.Errorf("failed to create admin role: %v", err2)
		}
		id, _ := res.LastInsertId()
		adminRoleID = int(id)
	}

	// Attribue le rôle admin à l'utilisateur
	_, err = common.DB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, adminRoleID)
	if err != nil {
		// Si l'insertion échoue, on vérifie si la liaison existe déjà
		var count int
		row = common.DB.QueryRow("SELECT COUNT(*) FROM user_roles WHERE user_id = ? AND role_id = ?", userID, adminRoleID)
		err2 := row.Scan(&count)
		if err2 != nil || count == 0 {
			return "", fmt.Errorf("failed to assign admin role: %v", err)
		}
		// La liaison existe déjà, c'est bon
	}

	// Login et récupère le token
	loginPayload := map[string]string{"email": email, "password": password}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	var jsonResp struct {
		Success bool `json:"success"`
		Data    struct {
			SessionToken string `json:"session_token"`
		} `json:"data"`
		Error string `json:"error"`
	}
	_ = json.Unmarshal(loginW.Body.Bytes(), &jsonResp)
	if !jsonResp.Success || jsonResp.Data.SessionToken == "" {
		return "", fmt.Errorf("login failed: %s", jsonResp.Error)
	}
	return jsonResp.Data.SessionToken, nil
}

// --- Helper local pour créer un utilisateur dans les tests de rôle ---
func createUserForRoleTest(router http.Handler, email, password, firstname, lastname string) *http.Response {
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

// --- Helper pour login et récupération de token (copié de user_test.go) ---
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
		Success      bool   `json:"success"`
		SessionToken string `json:"data.session_token"`
		Data         struct {
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

// --- Table-driven exhaustive pour toutes les routes du package role ---
func TestRole_AllRoutes_TableDriven(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()

	t.Run("POST /roles (création)", func(t *testing.T) {
		// Création d'un admin et récupération du token
		adminEmail := fmt.Sprintf("admin.role+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "Role")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		uniqueRoleName := fmt.Sprintf("moderator_%d", time.Now().UnixNano())

		cases := []struct {
			CaseName        string
			Token           string
			Payload         string
			PreCreateRole   bool
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Création rôle",
				Token:           token,
				Payload:         fmt.Sprintf(`{"name":"%s","description":"Rôle modérateur"}`, uniqueRoleName),
				PreCreateRole:   false,
				ExpectedStatus:  201,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessCreateRole,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				Payload:         fmt.Sprintf(`{"name":"unauth_%d","description":"Rôle"}`, time.Now().UnixNano()),
				PreCreateRole:   false,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName:        "Erreur - Données invalides (nom manquant)",
				Token:           token,
				Payload:         `{"description":"Pas de nom"}`,
				PreCreateRole:   false,
				ExpectedStatus:  400,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrInvalidData,
			},
			{
				CaseName:        "Erreur - Rôle déjà existant",
				Token:           token,
				Payload:         fmt.Sprintf(`{"name":"%s","description":"Rôle modérateur"}`, uniqueRoleName),
				PreCreateRole:   true,
				ExpectedStatus:  409,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrRoleAlreadyExists,
			},
		}

		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				if testCase.PreCreateRole {
					req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(testCase.Payload)))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
				}
				req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(testCase.Payload)))
				req.Header.Set("Content-Type", "application/json")
				if testCase.Token != "" {
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
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
	})

	// --- GET /roles (liste) ---
	t.Run("GET /roles (liste)", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin.list+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "List")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Créer quelques rôles
		for i := 0; i < 2; i++ {
			roleName := fmt.Sprintf("rolelist_%d_%d", time.Now().UnixNano(), i)
			payload := fmt.Sprintf(`{"name":"%s","description":"desc"}`, roleName)
			req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payload)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		cases := []struct {
			CaseName        string
			Token           string
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Liste des rôles",
				Token:           token,
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
		}
		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/roles", nil)
				if testCase.Token != "" {
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
				}
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				resp := w.Result()
				defer resp.Body.Close()
				var jsonResp struct {
					Success bool                     `json:"success"`
					Error   string                   `json:"error"`
					Data    []map[string]interface{} `json:"data"`
				}
				body, _ := io.ReadAll(resp.Body)
				_ = json.Unmarshal(body, &jsonResp)
				require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
				require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
				if testCase.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, testCase.ExpectedError)
				}
			})
		}
	})

	// --- GET /roles/:id ---
	t.Run("GET /roles/:id", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin.getid+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "GetID")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Créer un rôle pour le test
		roleName := fmt.Sprintf("getid_%d", time.Now().UnixNano())
		payload := fmt.Sprintf(`{"name":"%s","description":"desc"}`, roleName)
		req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var jsonResp struct {
			Success bool                   `json:"success"`
			Data    map[string]interface{} `json:"data"`
			Error   string                 `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &jsonResp)
		var roleID int
		if jsonResp.Data != nil {
			if id, ok := jsonResp.Data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}

		cases := []struct {
			CaseName        string
			Token           string
			ID              int
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Get rôle par ID",
				Token:           token,
				ID:              roleID,
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				ID:              roleID,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName:        "Erreur - Rôle inexistant",
				Token:           token,
				ID:              999999,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrRoleNotFound,
			},
		}
		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				url := fmt.Sprintf("/roles/%d", testCase.ID)
				req := httptest.NewRequest("GET", url, nil)
				if testCase.Token != "" {
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
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
				require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
				require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
				if testCase.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, testCase.ExpectedError)
				}
			})
		}
	})

	// --- PUT /roles/:id ---
	t.Run("PUT /roles/:id", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin.update+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "Update")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Créer un rôle pour le test
		roleName := fmt.Sprintf("update_%d", time.Now().UnixNano())
		payload := fmt.Sprintf(`{"name":"%s","description":"desc"}`, roleName)
		req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var jsonResp struct {
			Success bool                   `json:"success"`
			Data    map[string]interface{} `json:"data"`
			Error   string                 `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &jsonResp)
		var roleID int
		if jsonResp.Data != nil {
			if id, ok := jsonResp.Data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}

		// Créer un autre rôle pour le test de conflit
		conflictName := fmt.Sprintf("conflict_%d", time.Now().UnixNano())
		payloadConflict := fmt.Sprintf(`{"name":"%s","description":"desc"}`, conflictName)
		req2 := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payloadConflict)))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		cases := []struct {
			CaseName        string
			Token           string
			ID              int
			Payload         string
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Update rôle",
				Token:           token,
				ID:              roleID,
				Payload:         `{"description":"desc modifiée"}`,
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				ID:              roleID,
				Payload:         `{"description":"desc modifiée"}`,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName:        "Erreur - Rôle inexistant",
				Token:           token,
				ID:              999999,
				Payload:         `{"description":"desc"}`,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrRoleNotFound,
			},
			{
				CaseName:        "Erreur - Conflit nom déjà utilisé",
				Token:           token,
				ID:              roleID,
				Payload:         fmt.Sprintf(`{"name":"%s"}`, conflictName),
				ExpectedStatus:  409,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrRoleAlreadyExists,
			},
			{
				CaseName:        "Erreur - Données invalides",
				Token:           token,
				ID:              roleID,
				Payload:         `{"name":12345}`,
				ExpectedStatus:  400,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrInvalidData,
			},
		}
		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				url := fmt.Sprintf("/roles/%d", testCase.ID)
				req := httptest.NewRequest("PUT", url, bytes.NewReader([]byte(testCase.Payload)))
				req.Header.Set("Content-Type", "application/json")
				if testCase.Token != "" {
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
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
				require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
				require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
				if testCase.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, testCase.ExpectedError)
				}
			})
		}
	})

	// --- DELETE /roles/:id ---
	t.Run("DELETE /roles/:id", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin.delete+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "Delete")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Créer un rôle pour le test
		roleName := fmt.Sprintf("delete_%d", time.Now().UnixNano())
		payload := fmt.Sprintf(`{"name":"%s","description":"desc"}`, roleName)
		req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var jsonResp struct {
			Success bool                   `json:"success"`
			Data    map[string]interface{} `json:"data"`
			Error   string                 `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &jsonResp)
		var roleID int
		if jsonResp.Data != nil {
			if id, ok := jsonResp.Data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}

		cases := []struct {
			CaseName        string
			Token           string
			ID              int
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Delete rôle",
				Token:           token,
				ID:              roleID,
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				ID:              roleID,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName:        "Erreur - Rôle inexistant",
				Token:           token,
				ID:              999999,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrRoleNotFound,
			},
			{
				CaseName:        "Erreur - Déjà supprimé",
				Token:           token,
				ID:              roleID,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrRoleNotFound,
			},
		}

		for _, testCase := range cases {
			// Pour le cas 'Déjà supprimé', on supprime d'abord le rôle une première fois
			if testCase.CaseName == "Erreur - Déjà supprimé" {
				url := fmt.Sprintf("/roles/%d", testCase.ID)
				req := httptest.NewRequest("DELETE", url, nil)
				req.Header.Set("Authorization", "Bearer "+testCase.Token)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			}
			url := fmt.Sprintf("/roles/%d", testCase.ID)
			req := httptest.NewRequest("DELETE", url, nil)
			if testCase.Token != "" {
				req.Header.Set("Authorization", "Bearer "+testCase.Token)
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
			require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
			require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
			if testCase.ExpectedError != "" {
				require.Contains(t, jsonResp.Error, testCase.ExpectedError)
			}
		}
	})

	// --- POST /roles/assign ---
	t.Run("POST /roles/assign", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin.assign+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "Assign")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Créer un user
		userEmail := fmt.Sprintf("user.assign+%d@test.com", time.Now().UnixNano())
		userPassword := "motdepasse123"
		_ = createUserForRoleTest(router, userEmail, userPassword, "User", "Assign")
		var userID int
		row := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", userEmail)
		_ = row.Scan(&userID)

		// Créer un rôle
		roleName := fmt.Sprintf("assign_%d", time.Now().UnixNano())
		payload := fmt.Sprintf(`{"name":"%s","description":"desc"}`, roleName)
		req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var jsonResp struct {
			Success bool                   `json:"success"`
			Data    map[string]interface{} `json:"data"`
			Error   string                 `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &jsonResp)
		var roleID int
		if jsonResp.Data != nil {
			if id, ok := jsonResp.Data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}

		cases := []struct {
			CaseName        string
			Token           string
			UserID          int
			RoleID          int
			PreAssign       bool
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Assign role to user",
				Token:           token,
				UserID:          userID,
				RoleID:          roleID,
				PreAssign:       false,
				ExpectedStatus:  201,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				UserID:          userID,
				RoleID:          roleID,
				PreAssign:       false,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName:        "Erreur - User inexistant",
				Token:           token,
				UserID:          999999,
				RoleID:          roleID,
				PreAssign:       false,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotFound,
			},
			{
				CaseName:        "Erreur - Rôle inexistant",
				Token:           token,
				UserID:          userID,
				RoleID:          999999,
				PreAssign:       false,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrRoleNotFound,
			},
			{
				CaseName:        "Erreur - Déjà attribué",
				Token:           token,
				UserID:          userID,
				RoleID:          roleID,
				PreAssign:       true,
				ExpectedStatus:  409,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrRoleAlreadyAssigned,
			},
			{
				CaseName:        "Erreur - Données invalides",
				Token:           token,
				UserID:          userID,
				RoleID:          roleID,
				PreAssign:       false,
				ExpectedStatus:  400,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrInvalidData,
			},
		}

		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				if testCase.PreAssign {
					payload := fmt.Sprintf(`{"user_id":%d,"role_id":%d}`, testCase.UserID, testCase.RoleID)
					req := httptest.NewRequest("POST", "/roles/assign", bytes.NewReader([]byte(payload)))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
				}
				var payload string
				if testCase.CaseName == "Erreur - Données invalides" {
					payload = `{"user_id":"abc","role_id":null}`
				} else {
					payload = fmt.Sprintf(`{"user_id":%d,"role_id":%d}`, testCase.UserID, testCase.RoleID)
				}
				req := httptest.NewRequest("POST", "/roles/assign", bytes.NewReader([]byte(payload)))
				req.Header.Set("Content-Type", "application/json")
				if testCase.Token != "" {
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
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
				require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
				require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
				if testCase.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, testCase.ExpectedError)
				}
			})
		}
	})

	// --- POST /roles/revoke ---
	t.Run("POST /roles/revoke", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin.revoke+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "Revoke")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Créer un user
		userEmail := fmt.Sprintf("user.revoke+%d@test.com", time.Now().UnixNano())
		userPassword := "motdepasse123"
		_ = createUserForRoleTest(router, userEmail, userPassword, "User", "Revoke")
		var userID int
		row := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", userEmail)
		_ = row.Scan(&userID)

		// Créer un rôle
		roleName := fmt.Sprintf("revoke_%d", time.Now().UnixNano())
		payload := fmt.Sprintf(`{"name":"%s","description":"desc"}`, roleName)
		req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var jsonResp struct {
			Success bool                   `json:"success"`
			Data    map[string]interface{} `json:"data"`
			Error   string                 `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &jsonResp)
		var roleID int
		if jsonResp.Data != nil {
			if id, ok := jsonResp.Data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}

		// Pré-assigner le rôle pour le test de révocation
		payloadAssign := fmt.Sprintf(`{"user_id":%d,"role_id":%d}`, userID, roleID)
		reqAssign := httptest.NewRequest("POST", "/roles/assign", bytes.NewReader([]byte(payloadAssign)))
		reqAssign.Header.Set("Content-Type", "application/json")
		reqAssign.Header.Set("Authorization", "Bearer "+token)
		wAssign := httptest.NewRecorder()
		router.ServeHTTP(wAssign, reqAssign)

		cases := []struct {
			CaseName        string
			Token           string
			UserID          int
			RoleID          int
			PreRevoke       bool
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Revoke role from user",
				Token:           token,
				UserID:          userID,
				RoleID:          roleID,
				PreRevoke:       false,
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				UserID:          userID,
				RoleID:          roleID,
				PreRevoke:       false,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName:        "Erreur - Non attribué",
				Token:           token,
				UserID:          userID,
				RoleID:          roleID,
				PreRevoke:       true,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   "Ce rôle n'est pas attribué à cet utilisateur",
			},
			{
				CaseName:        "Erreur - Données invalides",
				Token:           token,
				UserID:          userID,
				RoleID:          roleID,
				PreRevoke:       false,
				ExpectedStatus:  400,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrInvalidData,
			},
		}

		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				if testCase.PreRevoke {
					// Révoquer une première fois pour simuler le cas "non attribué"
					payload := fmt.Sprintf(`{"user_id":%d,"role_id":%d}`, testCase.UserID, testCase.RoleID)
					req := httptest.NewRequest("POST", "/roles/revoke", bytes.NewReader([]byte(payload)))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
				}
				var payload string
				if testCase.CaseName == "Erreur - Données invalides" {
					payload = `{"user_id":null,"role_id":"abc"}`
				} else {
					payload = fmt.Sprintf(`{"user_id":%d,"role_id":%d}`, testCase.UserID, testCase.RoleID)
				}
				req := httptest.NewRequest("POST", "/roles/revoke", bytes.NewReader([]byte(payload)))
				req.Header.Set("Content-Type", "application/json")
				if testCase.Token != "" {
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
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
				require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
				require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
				if testCase.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, testCase.ExpectedError)
				}
			})
		}
	})

	// --- GET /roles/user/:user_id ---
	t.Run("GET /roles/user/:user_id", func(t *testing.T) {
		adminEmail := fmt.Sprintf("admin.getuserroles+%d@test.com", time.Now().UnixNano())
		adminPassword := "motdepasse123"
		token, err := createAdmin(router, adminEmail, adminPassword, "Admin", "GetUserRoles")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Créer un user
		userEmail := fmt.Sprintf("user.getuserroles+%d@test.com", time.Now().UnixNano())
		userPassword := "motdepasse123"
		_ = createUserForRoleTest(router, userEmail, userPassword, "User", "GetUserRoles")
		var userID int
		row := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", userEmail)
		_ = row.Scan(&userID)

		// Créer un rôle et l'assigner
		roleName := fmt.Sprintf("getuserroles_%d", time.Now().UnixNano())
		payload := fmt.Sprintf(`{"name":"%s","description":"desc"}`, roleName)
		req := httptest.NewRequest("POST", "/roles", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var jsonResp struct {
			Success bool                   `json:"success"`
			Data    map[string]interface{} `json:"data"`
			Error   string                 `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &jsonResp)
		var roleID int
		if jsonResp.Data != nil {
			if id, ok := jsonResp.Data["role_id"]; ok {
				roleID = int(id.(float64))
			}
		}
		payloadAssign := fmt.Sprintf(`{"user_id":%d,"role_id":%d}`, userID, roleID)
		reqAssign := httptest.NewRequest("POST", "/roles/assign", bytes.NewReader([]byte(payloadAssign)))
		reqAssign.Header.Set("Content-Type", "application/json")
		reqAssign.Header.Set("Authorization", "Bearer "+token)
		wAssign := httptest.NewRecorder()
		router.ServeHTTP(wAssign, reqAssign)

		cases := []struct {
			CaseName        string
			Token           string
			UserID          int
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedError   string
		}{
			{
				CaseName:        "Succès - Get user roles",
				Token:           token,
				UserID:          userID,
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedError:   "",
			},
			{
				CaseName:        "Erreur - Non authentifié",
				Token:           "",
				UserID:          userID,
				ExpectedStatus:  401,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotAuthenticated,
			},
			{
				CaseName:        "Erreur - User inexistant",
				Token:           token,
				UserID:          999999,
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedError:   common.ErrUserNotFound,
			},
		}
		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				url := fmt.Sprintf("/roles/user/%d", testCase.UserID)
				req := httptest.NewRequest("GET", url, nil)
				if testCase.Token != "" {
					req.Header.Set("Authorization", "Bearer "+testCase.Token)
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
				require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
				require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
				if testCase.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, testCase.ExpectedError)
				}
			})
		}
	})
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
