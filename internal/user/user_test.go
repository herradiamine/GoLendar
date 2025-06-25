package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/middleware"
	"go-averroes/internal/session"
	"go-averroes/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go-averroes/internal/common"

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

// --- Helpers mutualisés (à compléter selon tes besoins) ---

// Retourne l'ID de l'utilisateur créé
func createUser(t *testing.T, router http.Handler, email, password, firstname, lastname string) (int, *http.Response) {
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
	// Récupérer l'ID créé
	var userID int
	row := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", email)
	_ = row.Scan(&userID)
	token, err := loginAndGetToken(router, email, password)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return userID, w.Result()
}

// Retourne l'ID de l'admin créé
func createAdmin(t *testing.T, router http.Handler, email, password, firstname, lastname string) (int, *http.Response) {
	userID, resp := createUser(t, router, email, password, firstname, lastname)
	// Vérifie si le rôle admin existe, sinon le crée
	var adminRoleID int
	row := common.DB.QueryRow("SELECT role_id FROM roles WHERE name = 'admin'")
	err := row.Scan(&adminRoleID)
	if err != nil {
		// Le rôle admin n'existe pas, on le crée
		res, err2 := common.DB.Exec("INSERT INTO roles (name, description) VALUES ('admin', 'Administrateur')")
		if err2 != nil {
			panic("Impossible de créer le rôle admin: " + err2.Error())
		}
		id, _ := res.LastInsertId()
		adminRoleID = int(id)
	}
	// Attribue le rôle admin à l'utilisateur
	_, err = common.DB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, adminRoleID)
	if err != nil {
		panic("Impossible d'attribuer le rôle admin à l'utilisateur: " + err.Error())
	}
	token, err := loginAndGetToken(router, email, password)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return userID, resp
}

// --- Helper pour login et récupération de token ---
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

// --- Squelette de tests table-driven pour toutes les routes du package user.go ---

func TestAddUser(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	// Préparation des cas à tester
	var TestCases = []struct {
		CaseName        string
		Payload         string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string // message attendu (succès)
		ExpectedError   string // erreur attendue (échec)
	}{
		{
			CaseName:        "Succès - Création utilisateur",
			Payload:         `{"email":"testadduser1@example.com","password":"azerty1","firstname":"Jean","lastname":"Dupont"}`,
			ExpectedStatus:  201,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessCreateUser,
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Email invalide",
			Payload:         `{"email":"notanemail","password":"azerty1","firstname":"Jean","lastname":"Dupont"}`,
			ExpectedStatus:  400,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrInvalidData,
		},
		{
			CaseName:        "Erreur - Mot de passe trop court",
			Payload:         `{"email":"testadduser2@example.com","password":"123","firstname":"Jean","lastname":"Dupont"}`,
			ExpectedStatus:  400,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrInvalidData,
		},
		{
			CaseName:        "Erreur - Email déjà utilisé",
			Payload:         `{"email":"testadduser3@example.com","password":"azerty1","firstname":"Jean","lastname":"Dupont"}`,
			ExpectedStatus:  409,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserAlreadyExists,
		},
	}

	// Préparer un utilisateur pour le cas de conflit
	_, _ = createUser(t, router, "testadduser3@example.com", "azerty1", "Jean", "Dupont")

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			req := httptest.NewRequest("POST", "/user", strings.NewReader(testCase.Payload))
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

func TestGetUserMe(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	// Création d'un utilisateur et login pour obtenir un token
	email := "getme@example.com"
	password := "azerty1"
	_, _ = createUser(t, router, email, password, "Jean", "Dupont")
	token, err := loginAndGetToken(router, email, password)
	require.NoError(t, err)

	var TestCases = []struct {
		CaseName        string
		Token           string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Get user me",
			Token:           token,
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     "",
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			req := httptest.NewRequest("GET", "/user/me", nil)
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
}

func TestUpdateUserMe(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	email := "updateme@example.com"
	password := "azerty1"
	_, _ = createUser(t, router, email, password, "Jean", "Dupont")
	token, err := loginAndGetToken(router, email, password)
	require.NoError(t, err)

	var TestCases = []struct {
		CaseName        string
		Token           string
		Payload         string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Update firstname",
			Token:           token,
			Payload:         `{"firstname":"Paul"}`,
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessUserUpdate,
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			Payload:         `{"firstname":"Paul"}`,
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName:        "Erreur - Email invalide",
			Token:           token,
			Payload:         `{"email":"notanemail"}`,
			ExpectedStatus:  400,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrInvalidEmailFormat,
		},
		{
			CaseName:        "Erreur - Password trop court",
			Token:           token,
			Payload:         `{"password":"123"}`,
			ExpectedStatus:  400,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrPasswordTooShort,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			req := httptest.NewRequest("PUT", "/user/me", strings.NewReader(testCase.Payload))
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
}

func TestDeleteUserMe(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	email := "deleteme@example.com"
	password := "azerty1"
	_, _ = createUser(t, router, email, password, "Jean", "Dupont")
	token, err := loginAndGetToken(router, email, password)
	require.NoError(t, err)

	var TestCases = []struct {
		CaseName        string
		Token           string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Delete user me",
			Token:           token,
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessUserDelete,
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			req := httptest.NewRequest("DELETE", "/user/me", nil)
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
}

func TestGetUserWithRolesMe(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	email := "rolesme@example.com"
	password := "azerty1"
	_, _ = createUser(t, router, email, password, "Jean", "Dupont")
	token, err := loginAndGetToken(router, email, password)
	require.NoError(t, err)

	var TestCases = []struct {
		CaseName        string
		Token           string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Get user with roles me",
			Token:           token,
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     "", // Pas de message utilisateur spécifique attendu
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			req := httptest.NewRequest("GET", "/auth/me", nil)
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
}

func TestGetUserByID_Admin(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	// Création d'un admin et d'un user cible
	adminEmail := "admingetid@example.com"
	adminPassword := "azerty1"
	userEmail := "usergetid@example.com"
	userPassword := "azerty1"
	_, _ = createAdmin(t, router, adminEmail, adminPassword, "Admin", "Root")
	userID, _ := createUser(t, router, userEmail, userPassword, "Jean", "Dupont")
	adminToken, err := loginAndGetToken(router, adminEmail, adminPassword)
	require.NoError(t, err)

	var TestCases = []struct {
		CaseName        string
		Token           string
		UserID          string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Admin get user by id",
			Token:           adminToken,
			UserID:          fmt.Sprintf("%d", userID),
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     "",
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			UserID:          fmt.Sprintf("%d", userID),
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName:        "Erreur - User inexistant",
			Token:           adminToken,
			UserID:          "99999",
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotFound,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			url := "/user/" + testCase.UserID
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

func TestUpdateUserByID_Admin(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	// Création d'un admin et d'un user cible
	adminEmail := "adminupdateid@example.com"
	adminPassword := "azerty1"
	userEmail := "userupdateid@example.com"
	userPassword := "azerty1"
	_, _ = createAdmin(t, router, adminEmail, adminPassword, "Admin", "Root")
	userID, _ := createUser(t, router, userEmail, userPassword, "Jean", "Dupont")
	adminToken, err := loginAndGetToken(router, adminEmail, adminPassword)
	require.NoError(t, err)

	// Préparer un utilisateur avec l'email de conflit pour le test "Erreur - Email déjà utilisé"
	_, _ = createUser(t, router, "conflictupdateid@example.com", "azerty1", "Jean", "Dupont")

	var TestCases = []struct {
		CaseName        string
		Token           string
		UserID          string
		Payload         string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Admin update user by id",
			Token:           adminToken,
			UserID:          fmt.Sprintf("%d", userID),
			Payload:         `{"firstname":"Paul"}`,
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessUserUpdate,
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			UserID:          fmt.Sprintf("%d", userID),
			Payload:         `{"firstname":"Paul"}`,
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName:        "Erreur - User inexistant",
			Token:           adminToken,
			UserID:          "99999",
			Payload:         `{"firstname":"Paul"}`,
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotFound,
		},
		{
			CaseName:        "Erreur - Email déjà utilisé",
			Token:           adminToken,
			UserID:          fmt.Sprintf("%d", userID),
			Payload:         `{"email":"conflictupdateid@example.com"}`,
			ExpectedStatus:  409,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserAlreadyExists,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			url := "/user/" + testCase.UserID
			req := httptest.NewRequest("PUT", url, strings.NewReader(testCase.Payload))
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
}

func TestDeleteUserByID_Admin(t *testing.T) {
	testutils.ResetTestDB()
	// Création d'un admin et d'un user cible pour les cas classiques
	adminEmail := "admindeleteid@example.com"
	adminPassword := "azerty1"
	userEmail := "userdeleteid@example.com"
	userPassword := "azerty1"
	_, _ = createAdmin(t, setupTestRouter(), adminEmail, adminPassword, "Admin", "Root")
	userID, _ := createUser(t, setupTestRouter(), userEmail, userPassword, "Jean", "Dupont")
	adminToken, err := loginAndGetToken(setupTestRouter(), adminEmail, adminPassword)
	require.NoError(t, err)

	var TestCases = []struct {
		CaseName        string
		Token           string
		UserID          string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Admin delete user by id",
			Token:           adminToken,
			UserID:          fmt.Sprintf("%d", userID),
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessUserDelete,
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			UserID:          fmt.Sprintf("%d", userID),
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Erreur - User inexistant",
			// On crée un nouvel admin et récupère son token juste avant ce test pour garantir une session valide
			Token: func() string {
				newAdminEmail := "admininexistant@example.com"
				newAdminPassword := "azerty1"
				_, _ = createAdmin(t, setupTestRouter(), newAdminEmail, newAdminPassword, "Admin", "Root")
				token, err := loginAndGetToken(setupTestRouter(), newAdminEmail, newAdminPassword)
				require.NoError(t, err)
				return token
			}(),
			UserID:          "99999",
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotFound,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			url := "/user/" + testCase.UserID
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

func TestGetUserWithRolesByID_Admin(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()
	// Création d'un admin et d'un user cible
	adminEmail := "admingetrolesid@example.com"
	adminPassword := "azerty1"
	userEmail := "usergetrolesid@example.com"
	userPassword := "azerty1"
	_, _ = createAdmin(t, router, adminEmail, adminPassword, "Admin", "Root")
	userID, _ := createUser(t, router, userEmail, userPassword, "Jean", "Dupont")
	adminToken, err := loginAndGetToken(router, adminEmail, adminPassword)
	require.NoError(t, err)

	var TestCases = []struct {
		CaseName        string
		Token           string
		UserID          string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		{
			CaseName:        "Succès - Admin get user with roles by id",
			Token:           adminToken,
			UserID:          fmt.Sprintf("%d", userID),
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     "", // Pas de message utilisateur spécifique attendu
			ExpectedError:   "",
		},
		{
			CaseName:        "Erreur - Non authentifié",
			Token:           "",
			UserID:          fmt.Sprintf("%d", userID),
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName:        "Erreur - User inexistant",
			Token:           adminToken,
			UserID:          "99999",
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotFound,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := setupTestRouter()
			url := "/user/" + testCase.UserID + "/with-roles"
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
