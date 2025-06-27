package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// TestLogin teste la route de connexion avec plusieurs cas
func TestLogin(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() any
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Connexion réussie avec utilisateur normal",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(false, true) // Sans session
				return map[string]interface{}{
					"email":    user.User.Email,
					"password": user.Password,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessLogin,
			ExpectedError:    "",
		},
		{
			CaseName: "Connexion réussie avec admin",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				admin, _ := testutils.GenerateAuthenticatedAdmin(false, true) // Sans session
				return map[string]interface{}{
					"email":    admin.User.Email,
					"password": admin.Password,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessLogin,
			ExpectedError:    "",
		},
		{
			CaseName: "Email manquant",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return map[string]interface{}{
					"password": "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Mot de passe manquant",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return map[string]interface{}{
					"email": "test@example.com",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Email invalide",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return map[string]interface{}{
					"email":    "invalid-email",
					"password": "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Email avec espaces",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return map[string]interface{}{
					"email":    " test@example.com ",
					"password": "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Email vide",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return map[string]interface{}{
					"email":    "",
					"password": "password123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Mot de passe vide",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return map[string]interface{}{
					"email":    "test@example.com",
					"password": "",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Utilisateur inexistant",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return map[string]interface{}{
					"email":    "nonexistent@example.com",
					"password": "password123",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCredentials,
		},
		{
			CaseName: "Mot de passe incorrect",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(false, true) // Sans session
				return map[string]interface{}{
					"email":    user.User.Email,
					"password": "wrongpassword",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCredentials,
		},
		{
			CaseName: "Données JSON invalides",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				return "invalid json string"
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Champs supplémentaires ignorés",
			CaseUrl:  "/auth/login",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(false, true) // Sans session
				return map[string]interface{}{
					"email":         user.User.Email,
					"password":      user.Password,
					"extra_field":   "should be ignored",
					"another_field": 123,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessLogin,
			ExpectedError:    "",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			testutils.PurgeAllTestUsers()

			// On prépare les données utiles au traitement de ce cas.
			router := testutils.CreateTestRouter()
			w := httptest.NewRecorder()

			// Récupérer les données de setup
			setupData := testCase.SetupData()

			// Gérer les différents types de données de setup
			var jsonData []byte
			var err error

			if requestData, ok := setupData.(map[string]interface{}); ok {
				jsonData, err = json.Marshal(requestData)
			} else if jsonString, ok := setupData.(string); ok {
				jsonData = []byte(jsonString)
			} else {
				t.Fatalf("Type de données de setup non supporté")
			}

			require.NoError(t, err)

			req, _ := http.NewRequest("POST", testCase.CaseUrl, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// On traite les cas de test un par un.
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

			// Vérifications spécifiques pour les connexions réussies
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				require.NotEmpty(t, response.Data)
				require.Contains(t, response.Data, "session_token")
				require.Contains(t, response.Data, "refresh_token")
				require.Contains(t, response.Data, "expires_at")
				require.Contains(t, response.Data, "user")
				require.Contains(t, response.Data, "roles")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestLogout teste la route de déconnexion
func TestLogout(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() any
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Déconnexion réussie avec token valide",
			CaseUrl:  "/auth/logout",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessLogout,
			ExpectedError:    "",
		},
		{
			CaseName: "Déconnexion sans token",
			CaseUrl:  "/auth/logout",
			SetupData: func() any {
				return ""
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Déconnexion avec token invalide",
			CaseUrl:  "/auth/logout",
			SetupData: func() any {
				return "Bearer invalid_token_12345"
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Déconnexion avec token expiré",
			CaseUrl:  "/auth/logout",
			SetupData: func() any {
				return "Bearer expired_token_12345"
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Déconnexion avec format de token incorrect",
			CaseUrl:  "/auth/logout",
			SetupData: func() any {
				return "InvalidFormat token_12345"
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Déconnexion avec token vide",
			CaseUrl:  "/auth/logout",
			SetupData: func() any {
				return "Bearer "
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Déconnexion avec token de session supprimée",
			CaseUrl:  "/auth/logout",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				// Supprimer la session en base
				common.DB.Exec("UPDATE user_session SET is_active = FALSE WHERE session_token = ?", user.SessionToken)
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			testutils.PurgeAllTestUsers()

			// On prépare les données utiles au traitement de ce cas.
			router := testutils.CreateTestRouter()
			w := httptest.NewRecorder()

			// Récupérer les données de setup
			authHeader := testCase.SetupData().(string)

			// Créer la requête
			req, _ := http.NewRequest("POST", testCase.CaseUrl, nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			// On traite les cas de test un par un.
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

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestRefreshToken teste la route de rafraîchissement de token
func TestRefreshToken(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() any
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Rafraîchissement réussi avec refresh token valide",
			CaseUrl:  "/auth/refresh",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				return map[string]interface{}{
					"refresh_token": user.RefreshToken,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessRefreshToken,
			ExpectedError:    "",
		},
		{
			CaseName: "Rafraîchissement avec refresh token invalide",
			CaseUrl:  "/auth/refresh",
			SetupData: func() any {
				return map[string]interface{}{
					"refresh_token": "invalid_refresh_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Rafraîchissement sans refresh token",
			CaseUrl:  "/auth/refresh",
			SetupData: func() any {
				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Rafraîchissement avec refresh token vide",
			CaseUrl:  "/auth/refresh",
			SetupData: func() any {
				return map[string]interface{}{
					"refresh_token": "",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Rafraîchissement avec refresh token expiré",
			CaseUrl:  "/auth/refresh",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				// Expirer la session en base
				common.DB.Exec("UPDATE user_session SET expires_at = DATE_SUB(NOW(), INTERVAL 1 HOUR) WHERE session_token = ?", user.SessionToken)
				return map[string]interface{}{
					"refresh_token": user.RefreshToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionExpired,
		},
		{
			CaseName: "Rafraîchissement avec session inactive",
			CaseUrl:  "/auth/refresh",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				// Désactiver la session en base
				common.DB.Exec("UPDATE user_session SET is_active = FALSE WHERE session_token = ?", user.SessionToken)
				return map[string]interface{}{
					"refresh_token": user.RefreshToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Rafraîchissement avec données JSON invalides",
			CaseUrl:  "/auth/refresh",
			SetupData: func() any {
				return "invalid json string"
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			testutils.PurgeAllTestUsers()

			// On prépare les données utiles au traitement de ce cas.
			router := testutils.CreateTestRouter()
			w := httptest.NewRecorder()

			// Récupérer les données de setup
			setupData := testCase.SetupData()

			// Gérer les différents types de données de setup
			var jsonData []byte
			var err error

			if requestData, ok := setupData.(map[string]interface{}); ok {
				jsonData, err = json.Marshal(requestData)
			} else if jsonString, ok := setupData.(string); ok {
				jsonData = []byte(jsonString)
			} else {
				t.Fatalf("Type de données de setup non supporté")
			}

			require.NoError(t, err)

			req, _ := http.NewRequest("POST", testCase.CaseUrl, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// On traite les cas de test un par un.
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

			// Vérifications spécifiques pour le rafraîchissement réussi
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				require.NotEmpty(t, response.Data)
				require.Contains(t, response.Data, "session_token")
				require.Contains(t, response.Data, "expires_at")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetUserSessions teste la récupération des sessions d'un utilisateur
func TestGetUserSessions(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() any
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération des sessions avec utilisateur authentifié",
			CaseUrl:  "/auth/sessions",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération des sessions sans authentification",
			CaseUrl:  "/auth/sessions",
			SetupData: func() any {
				return ""
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Récupération des sessions avec token invalide",
			CaseUrl:  "/auth/sessions",
			SetupData: func() any {
				return "Bearer invalid_token_12345"
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Récupération des sessions avec utilisateur ayant plusieurs sessions",
			CaseUrl:  "/auth/sessions",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Première session
				// Créer une deuxième session pour le même utilisateur
				_, _, _, _ = testutils.CreateUserSession(user.User.UserID, 24*time.Hour)
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération des sessions avec utilisateur sans sessions actives",
			CaseUrl:  "/auth/sessions",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				// Supprimer toutes les sessions de l'utilisateur
				common.DB.Exec("UPDATE user_session SET deleted_at = NOW() WHERE user_id = ?", user.User.UserID)
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			testutils.PurgeAllTestUsers()

			// On prépare les données utiles au traitement de ce cas.
			router := testutils.CreateTestRouter()
			w := httptest.NewRecorder()

			// Récupérer les données de setup
			authHeader := testCase.SetupData().(string)

			// Créer la requête
			req, _ := http.NewRequest("GET", testCase.CaseUrl, nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			// On traite les cas de test un par un.
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

			// Vérifications spécifiques pour la récupération réussie
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				require.NotEmpty(t, response.Data)
				// Vérifier que les sessions sont retournées
				sessions, ok := response.Data.([]interface{})
				require.True(t, ok)
				require.GreaterOrEqual(t, len(sessions), 1)
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestDeleteSession teste la suppression d'une session spécifique
func TestDeleteSession(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() any
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Suppression réussie d'une session valide",
			CaseUrl:  "/auth/sessions/1",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteSession,
			ExpectedError:    "",
		},
		{
			CaseName: "Suppression sans authentification",
			CaseUrl:  "/auth/sessions/1",
			SetupData: func() any {
				return ""
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Suppression avec session ID invalide",
			CaseUrl:  "/auth/sessions/99999",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionNotFound,
		},
		{
			CaseName: "Suppression avec session ID non numérique",
			CaseUrl:  "/auth/sessions/abc",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionNotFound,
		},
		{
			CaseName: "Suppression d'une session qui ne vous appartient pas",
			CaseUrl:  "/auth/sessions/1",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				// Créer deux utilisateurs
				user1, _ := testutils.GenerateAuthenticatedUser(true, true)
				user2, _ := testutils.GenerateAuthenticatedUser(true, true)
				// Utiliser le token de user1 mais essayer de supprimer la session de user2
				// Récupérer l'ID de session de user2 pour le test
				var sessionID int
				common.DB.QueryRow("SELECT user_session_id FROM user_session WHERE user_id = ? LIMIT 1", user2.User.UserID).Scan(&sessionID)
				// Modifier l'URL pour utiliser l'ID de session de user2
				return map[string]interface{}{
					"authHeader": "Bearer " + user1.SessionToken,
					"sessionID":  sessionID,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionNotFound,
		},
		{
			CaseName: "Suppression d'une session déjà supprimée",
			CaseUrl:  "/auth/sessions/1",
			SetupData: func() any {
				testutils.PurgeAllTestUsers()
				user, _ := testutils.GenerateAuthenticatedUser(true, true) // Avec session
				// Supprimer la session en base
				common.DB.Exec("UPDATE user_session SET deleted_at = NOW() WHERE user_id = ?", user.User.UserID)
				return "Bearer " + user.SessionToken
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Suppression avec token invalide",
			CaseUrl:  "/auth/sessions/1",
			SetupData: func() any {
				return "Bearer invalid_token_12345"
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			testutils.PurgeAllTestUsers()

			// On prépare les données utiles au traitement de ce cas.
			router := testutils.CreateTestRouter()
			w := httptest.NewRecorder()

			// Récupérer les données de setup
			setupData := testCase.SetupData()

			var authHeader string
			var sessionID string

			// Gérer les différents types de données de setup
			if authStr, ok := setupData.(string); ok {
				authHeader = authStr
				sessionID = "1" // Valeur par défaut
			} else if setupMap, ok := setupData.(map[string]interface{}); ok {
				if auth, exists := setupMap["authHeader"]; exists {
					authHeader = auth.(string)
				}
				if session, exists := setupMap["sessionID"]; exists {
					sessionID = fmt.Sprintf("%d", session.(int))
				} else {
					sessionID = "1"
				}
			}

			// Créer la requête avec l'URL dynamique si nécessaire
			url := testCase.CaseUrl
			if sessionID != "1" {
				url = "/auth/sessions/" + sessionID
			}

			req, _ := http.NewRequest("DELETE", url, nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			// On traite les cas de test un par un.
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

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}
