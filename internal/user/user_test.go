package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-averroes/internal/common"
	"go-averroes/testutils"

	"github.com/stretchr/testify/require"
)

// TestMain configure l'environnement de test global
func TestMain(m *testing.M) {
	// Initialiser l'environnement de test
	if err := testutils.SetupTestEnvironment(); err != nil {
		panic("Impossible d'initialiser l'environnement de test: " + err.Error())
	}

	// Exécuter les tests
	code := m.Run()

	// Nettoyer l'environnement de test
	if err := testutils.TeardownTestEnvironment(); err != nil {
		panic("Impossible de nettoyer l'environnement de test: " + err.Error())
	}

	// Retourner le code de sortie
	os.Exit(code)
}

// TestUserAdd teste la route POST /user (création d'utilisateur)
func TestUserAdd(t *testing.T) {
	router := testutils.CreateTestRouter()

	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Création d'utilisateur avec données valides",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Lastname:  "Dupont",
					Firstname: "Jean",
					Email:     "jean.dupont@example.com",
					Password:  "password123",
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Données JSON invalides",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return `{"invalid": "json"`
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Email manquant",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Lastname:  "Dupont",
					Firstname: "Jean",
					Password:  "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Prénom manquant",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Lastname: "Dupont",
					Email:    "jean.dupont@example.com",
					Password: "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Nom manquant",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Firstname: "Jean",
					Email:     "jean.dupont@example.com",
					Password:  "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Mot de passe manquant",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Lastname:  "Dupont",
					Firstname: "Jean",
					Email:     "jean.dupont@example.com",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Mot de passe trop court",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Lastname:  "Dupont",
					Firstname: "Jean",
					Email:     "jean.dupont@example.com",
					Password:  "123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Format d'email invalide",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Lastname:  "Dupont",
					Firstname: "Jean",
					Email:     "email-invalide",
					Password:  "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Email avec espaces",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				return common.CreateUserRequest{
					Lastname:  "Dupont",
					Firstname: "Jean",
					Email:     " jean.dupont@example.com ",
					Password:  "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Email déjà existant",
			CaseUrl:  "/user",
			SetupData: func() interface{} {
				// Créer d'abord un utilisateur
				user, err := testutils.GenerateAuthenticatedUser(false)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				return common.CreateUserRequest{
					Lastname:  "Autre",
					Firstname: "Utilisateur",
					Email:     user.User.Email,
					Password:  "password123",
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			data := testCase.SetupData()

			// Préparer la requête
			var body []byte
			var err error

			if str, ok := data.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(data)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", testCase.CaseUrl, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			// Vérifications spécifiques pour la création réussie
			if testCase.ExpectedHttpCode == http.StatusCreated {
				require.True(t, response.Success)
				require.NotNil(t, response.Data)

				// Vérifier que l'utilisateur a bien été créé en base
				if userData, ok := response.Data.(map[string]interface{}); ok {
					userID := int(userData["user_id"].(float64))
					require.Greater(t, userID, 0)
				}
			}
		})
	}
}

// TestUserGetMe teste la route GET /user/me
func TestUserGetMe(t *testing.T) {
	router := testutils.CreateTestRouter()

	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() string
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié récupère ses données",
			CaseUrl:  "/user/me",
			SetupData: func() string {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié",
			CaseUrl:  "/user/me",
			SetupData: func() string {
				return ""
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Token invalide",
			CaseUrl:  "/user/me",
			SetupData: func() string {
				return "Bearer invalid-token"
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer la requête
			req := httptest.NewRequest("GET", testCase.CaseUrl, nil)

			// Ajouter le token d'authentification si fourni
			if token := testCase.SetupData(); token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				require.NotNil(t, response.Data)
			}
		})
	}
}

// TestUserUpdateMe teste la route PUT /user/me
func TestUserUpdateMe(t *testing.T) {
	router := testutils.CreateTestRouter()

	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() (string, interface{})
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Mise à jour du prénom uniquement",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				newFirstname := "NouveauPrénom"
				return "Bearer " + user.SessionToken, common.UpdateUserRequest{
					Firstname: &newFirstname,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour du nom uniquement",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				newLastname := "NouveauNom"
				return "Bearer " + user.SessionToken, common.UpdateUserRequest{
					Lastname: &newLastname,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour de l'email uniquement",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				newEmail := "nouveau.email@example.com"
				return "Bearer " + user.SessionToken, common.UpdateUserRequest{
					Email: &newEmail,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour du mot de passe uniquement",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				newPassword := "nouveauMotDePasse123"
				return "Bearer " + user.SessionToken, common.UpdateUserRequest{
					Password: &newPassword,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour de plusieurs champs",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				newFirstname := "Prénom"
				newLastname := "Nom"
				newEmail := "prenom.nom@example.com"
				return "Bearer " + user.SessionToken, common.UpdateUserRequest{
					Firstname: &newFirstname,
					Lastname:  &newLastname,
					Email:     &newEmail,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				return "", common.UpdateUserRequest{
					Firstname: stringPtr("Nouveau"),
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Données JSON invalides",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				return "Bearer " + user.SessionToken, `{"invalid": "json"`
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Format d'email invalide",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				invalidEmail := "email-invalide"
				return "Bearer " + user.SessionToken, common.UpdateUserRequest{
					Email: &invalidEmail,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEmailFormat,
		},
		{
			CaseName: "Mot de passe trop court",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				shortPassword := "123"
				return "Bearer " + user.SessionToken, common.UpdateUserRequest{
					Password: &shortPassword,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrPasswordTooShort,
		},
		{
			CaseName: "Email déjà utilisé par un autre utilisateur",
			CaseUrl:  "/user/me",
			SetupData: func() (string, interface{}) {
				// Créer deux utilisateurs
				user1, err1 := testutils.GenerateAuthenticatedUser(true)
				if err1 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err1.Error())
				}
				user2, err2 := testutils.GenerateAuthenticatedUser(false)
				if err2 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err2.Error())
				}

				// User1 essaie d'utiliser l'email de user2
				return "Bearer " + user1.SessionToken, common.UpdateUserRequest{
					Email: &user2.User.Email,
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			token, data := testCase.SetupData()

			// Préparer la requête
			var body []byte
			var err error

			if str, ok := data.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(data)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", testCase.CaseUrl, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
			}
		})
	}
}

// TestUserDeleteMe teste la route DELETE /user/me
func TestUserDeleteMe(t *testing.T) {
	router := testutils.CreateTestRouter()

	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() string
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié supprime son compte",
			CaseUrl:  "/user/me",
			SetupData: func() string {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserDelete,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié",
			CaseUrl:  "/user/me",
			SetupData: func() string {
				return ""
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer la requête
			req := httptest.NewRequest("DELETE", testCase.CaseUrl, nil)

			// Ajouter le token d'authentification si fourni
			if token := testCase.SetupData(); token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
			}
		})
	}
}

// TestAuthMe teste la route GET /auth/me
func TestAuthMe(t *testing.T) {
	router := testutils.CreateTestRouter()

	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() string
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur récupéré avec ses rôles",
			CaseUrl:  "/auth/me",
			SetupData: func() string {
				user, err := testutils.GenerateAuthenticatedAdmin(true)
				if err != nil {
					panic("Erreur lors de la création de l'admin de test: " + err.Error())
				}
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur sans rôles",
			CaseUrl:  "/auth/me",
			SetupData: func() string {
				user, err := testutils.GenerateAuthenticatedUser(true)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié",
			CaseUrl:  "/auth/me",
			SetupData: func() string {
				return ""
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer la requête
			req := httptest.NewRequest("GET", testCase.CaseUrl, nil)

			// Ajouter le token d'authentification si fourni
			if token := testCase.SetupData(); token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				require.NotNil(t, response.Data)

				// Vérifier que les données contiennent l'utilisateur et ses rôles
				if userWithRoles, ok := response.Data.(map[string]interface{}); ok {
					require.Contains(t, userWithRoles, "user")
					require.Contains(t, userWithRoles, "roles")
				}
			}
		})
	}
}

// TestUserAdminRoutes teste les routes admin pour la gestion des utilisateurs
func TestUserAdminRoutes(t *testing.T) {
	router := testutils.CreateTestRouter()

	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() (string, int)
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Admin récupère un utilisateur par ID",
			CaseUrl:  "/user/%d",
			SetupData: func() (string, int) {
				admin, err1 := testutils.GenerateAuthenticatedAdmin(true)
				if err1 != nil {
					panic("Erreur lors de la création de l'admin de test: " + err1.Error())
				}
				user, err2 := testutils.GenerateAuthenticatedUser(false)
				if err2 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err2.Error())
				}
				return "Bearer " + admin.SessionToken, user.User.UserID
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Admin met à jour un utilisateur par ID",
			CaseUrl:  "/user/%d",
			SetupData: func() (string, int) {
				admin, err1 := testutils.GenerateAuthenticatedAdmin(true)
				if err1 != nil {
					panic("Erreur lors de la création de l'admin de test: " + err1.Error())
				}
				user, err2 := testutils.GenerateAuthenticatedUser(false)
				if err2 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err2.Error())
				}
				return "Bearer " + admin.SessionToken, user.User.UserID
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin supprime un utilisateur par ID",
			CaseUrl:  "/user/%d",
			SetupData: func() (string, int) {
				admin, err1 := testutils.GenerateAuthenticatedAdmin(true)
				if err1 != nil {
					panic("Erreur lors de la création de l'admin de test: " + err1.Error())
				}
				user, err2 := testutils.GenerateAuthenticatedUser(false)
				if err2 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err2.Error())
				}
				return "Bearer " + admin.SessionToken, user.User.UserID
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserDelete,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin récupère un utilisateur avec ses rôles",
			CaseUrl:  "/user/%d/with-roles",
			SetupData: func() (string, int) {
				admin, err1 := testutils.GenerateAuthenticatedAdmin(true)
				if err1 != nil {
					panic("Erreur lors de la création de l'admin de test: " + err1.Error())
				}
				user, err2 := testutils.GenerateAuthenticatedUser(false)
				if err2 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err2.Error())
				}
				return "Bearer " + admin.SessionToken, user.User.UserID
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non admin tente d'accéder",
			CaseUrl:  "/user/%d",
			SetupData: func() (string, int) {
				user, err1 := testutils.GenerateAuthenticatedUser(true)
				if err1 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err1.Error())
				}
				otherUser, err2 := testutils.GenerateAuthenticatedUser(false)
				if err2 != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err2.Error())
				}
				return "Bearer " + user.SessionToken, otherUser.User.UserID
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Utilisateur non authentifié",
			CaseUrl:  "/user/%d",
			SetupData: func() (string, int) {
				user, err := testutils.GenerateAuthenticatedUser(false)
				if err != nil {
					panic("Erreur lors de la création de l'utilisateur de test: " + err.Error())
				}
				return "", user.User.UserID
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "ID utilisateur invalide",
			CaseUrl:  "/user/invalid",
			SetupData: func() (string, int) {
				admin, err := testutils.GenerateAuthenticatedAdmin(true)
				if err != nil {
					panic("Erreur lors de la création de l'admin de test: " + err.Error())
				}
				return "Bearer " + admin.SessionToken, 0
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidUserID,
		},
		{
			CaseName: "Utilisateur inexistant",
			CaseUrl:  "/user/%d",
			SetupData: func() (string, int) {
				admin, err := testutils.GenerateAuthenticatedAdmin(true)
				if err != nil {
					panic("Erreur lors de la création de l'admin de test: " + err.Error())
				}
				return "Bearer " + admin.SessionToken, 99999
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			token, userID := testCase.SetupData()

			// Construire l'URL
			var url string
			if testCase.CaseUrl == "/user/invalid" {
				url = "/user/invalid"
			} else {
				url = fmt.Sprintf(testCase.CaseUrl, userID)
			}

			// Déterminer la méthode HTTP basée sur le message attendu
			method := "GET"
			if testCase.ExpectedMessage == common.MsgSuccessUserUpdate {
				method = "PUT"
			} else if testCase.ExpectedMessage == common.MsgSuccessUserDelete {
				method = "DELETE"
			}

			// Préparer la requête
			var req *http.Request
			if method == "PUT" {
				// Pour PUT, on envoie des données de mise à jour
				updateData := common.UpdateUserRequest{
					Firstname: stringPtr("NouveauPrénom"),
				}
				body, _ := json.Marshal(updateData)
				req = httptest.NewRequest(method, url, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(method, url, nil)
			}

			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
			}
		})
	}
}

// Fonction utilitaire pour créer un pointeur vers une string
func stringPtr(s string) *string {
	return &s
}
