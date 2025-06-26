package session_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-averroes/testutils"

	"github.com/gin-gonic/gin"
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

// TestRouteExample teste la route d'exemple avec plusieurs cas
func TestRouteExample(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName:         "Case name",
			CaseUrl:          "Url",
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Success message",
			ExpectedError:    "Error message",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			// On traite les cas de test un par un.
			require.Equal(t, testCase.CaseUrl, "Url")
			require.Equal(t, testCase.ExpectedHttpCode, http.StatusOK)
			require.Equal(t, testCase.ExpectedMessage, "Success message")
			require.Equal(t, testCase.ExpectedError, "Error message")
			// On purge les données après avoir traité le cas.
		})
	}
}

// TestGetUserSessions teste la récupération des sessions utilisateur via /auth/sessions
func TestGetUserSessions(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié - sessions récupérées",
			SetupAuth: func() (string, string, error) {
				// Utilise une fonction utilitaire pour créer un utilisateur et récupérer un token
				// (à adapter selon tes utilitaires dans testutils)
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", testutils.GenerateUniqueEmail("jean.dupont"))
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - accès refusé",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    "Utilisateur non authentifié", // À adapter selon ton message d'erreur
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/auth/sessions", nil)
			require.NoError(t, err)

			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response["error"], testCase.ExpectedError)
			}

			// Nettoyage
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestLogout teste la déconnexion via /auth/logout
func TestLogout(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié - déconnexion réussie",
			SetupAuth: func() (string, string, error) {
				user, token, err := testutils.CreateAuthenticatedUser(2, "Martin", "Paul", testutils.GenerateUniqueEmail("paul.martin"))
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - accès refusé",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    "Utilisateur non authentifié", // À adapter
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/auth/logout", nil)
			require.NoError(t, err)

			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response["error"], testCase.ExpectedError)
			}

			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestGetAuthMe teste la récupération des infos utilisateur via /auth/me
func TestGetAuthMe(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié - infos récupérées",
			SetupAuth: func() (string, string, error) {
				user, token, err := testutils.CreateAuthenticatedUser(3, "Durand", "Alice", testutils.GenerateUniqueEmail("alice.durand"))
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - accès refusé",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    "Utilisateur non authentifié", // À adapter
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response["error"], testCase.ExpectedError)
			}

			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestDeleteSession teste la suppression d'une session via /auth/sessions/:session_id
func TestDeleteSession(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (token string, userEmail string, sessionID string, err error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié - suppression de session existante",
			SetupAuth: func() (string, string, string, error) {
				user, token, err := testutils.CreateAuthenticatedUser(4, "Lefevre", "Luc", testutils.GenerateUniqueEmail("luc.lefevre"))
				if err != nil {
					return "", "", "", err
				}
				// Récupérer la session courante de l'utilisateur (à adapter selon ta structure)
				sessionID, err := testutils.GetSessionIDForUser(user.UserID)
				if err != nil {
					return "", "", "", err
				}
				return "Bearer " + token, user.Email, sessionID, nil
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - accès refusé",
			SetupAuth: func() (string, string, string, error) {
				return "", "", "fake-session-id", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    "Utilisateur non authentifié", // À adapter
		},
		{
			CaseName: "Session inexistante - erreur",
			SetupAuth: func() (string, string, string, error) {
				user, token, err := testutils.CreateAuthenticatedUser(5, "Moreau", "Emma", testutils.GenerateUniqueEmail("emma.moreau"))
				if err != nil {
					return "", "", "", err
				}
				return "Bearer " + token, user.Email, "session-inexistante", nil
			},
			ExpectedHttpCode: http.StatusNotFound,   // À adapter selon ton code
			ExpectedError:    "session non trouvée", // À adapter selon ton message d'erreur
		},
		{
			CaseName: "Suppression d'une session appartenant à un autre utilisateur",
			SetupAuth: func() (string, string, string, error) {
				// Créer deux utilisateurs
				user1, token1, err := testutils.CreateAuthenticatedUser(6, "Martin", "Jean", testutils.GenerateUniqueEmail("jean.martin"))
				if err != nil {
					return "", "", "", err
				}
				user2, _, err := testutils.CreateAuthenticatedUser(7, "Durand", "Paul", testutils.GenerateUniqueEmail("paul.durand"))
				if err != nil {
					return "", "", "", err
				}
				// Récupérer la session de user2
				sessionID, err := testutils.GetUserSessionIDForUser(user2.UserID)
				if err != nil {
					return "", "", "", err
				}
				return "Bearer " + token1, user1.Email, sessionID, nil
			},
			ExpectedHttpCode: http.StatusNotFound,   // ou 403 selon ta politique
			ExpectedError:    "session non trouvée", // ou message adapté
		},
		{
			CaseName: "Suppression d'une session déjà supprimée",
			SetupAuth: func() (string, string, string, error) {
				user, token, err := testutils.CreateAuthenticatedUser(8, "Petit", "Lucie", testutils.GenerateUniqueEmail("lucie.petit"))
				if err != nil {
					return "", "", "", err
				}
				sessionID, err := testutils.GetUserSessionIDForUser(user.UserID)
				if err != nil {
					return "", "", "", err
				}
				// Supprimer la session une première fois
				err = testutils.DeleteUserSessionByID(sessionID)
				if err != nil {
					return "", "", "", err
				}
				// Tenter de la supprimer à nouveau
				return "Bearer " + token, user.Email, sessionID, nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    "Session invalide",
		},
		{
			CaseName: "Suppression d'une session secondaire (multi-session)",
			SetupAuth: func() (string, string, string, error) {
				user, token, err := testutils.CreateAuthenticatedUser(9, "Lemoine", "Sophie", testutils.GenerateUniqueEmail("sophie.lemoine"))
				if err != nil {
					return "", "", "", err
				}
				// Créer une deuxième session pour le même utilisateur
				secondToken, err := testutils.CreateSessionForUser(user.UserID)
				if err != nil {
					return "", "", "", err
				}
				secondSessionID, err := testutils.GetUserSessionIDByToken(secondToken)
				if err != nil {
					return "", "", "", err
				}
				return "Bearer " + token, user.Email, secondSessionID, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Vérification post-suppression : accès avec le même token retourne 401",
			SetupAuth: func() (string, string, string, error) {
				user, token, err := testutils.CreateAuthenticatedUser(10, "Fabre", "Julie", testutils.GenerateUniqueEmail("julie.fabre"))
				if err != nil {
					return "", "", "", err
				}
				sessionID, err := testutils.GetUserSessionIDForUser(user.UserID)
				if err != nil {
					return "", "", "", err
				}
				return "Bearer " + token, user.Email, sessionID, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userEmail, sessionID, err := testCase.SetupAuth()
			require.NoError(t, err)

			url := "/auth/sessions/" + sessionID
			req, err := http.NewRequest("DELETE", url, nil)
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response["error"], testCase.ExpectedError)
			}

			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestLogin teste la route POST /auth/login
func TestLogin(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		RequestData      map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		SetupUser        func() (email, password string, cleanup func())
	}{
		{
			CaseName: "Login succès (identifiants valides)",
			RequestData: map[string]interface{}{
				"email":    "login.succes@example.com",
				"password": "password123",
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Connexion réussie", // À adapter selon ton message
			ExpectedError:    "",
			SetupUser: func() (string, string, func()) {
				email := testutils.GenerateUniqueEmail("login.succes")
				_, err := testutils.CreateUserWithPassword("Test", "User", email, "password123")
				cleanup := func() { _ = testutils.PurgeTestData(email) }
				if err != nil {
					return "", "", cleanup
				}
				return email, "password123", cleanup
			},
		},
		{
			CaseName: "Mot de passe incorrect",
			RequestData: map[string]interface{}{
				"email":    "login.badpass@example.com",
				"password": "wrongpassword",
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "Email ou mot de passe incorrect", // À adapter
			SetupUser: func() (string, string, func()) {
				email := testutils.GenerateUniqueEmail("login.badpass")
				_, err := testutils.CreateUserWithPassword("Test", "User", email, "password123")
				cleanup := func() { _ = testutils.PurgeTestData(email) }
				if err != nil {
					return "", "", cleanup
				}
				return email, "password123", cleanup
			},
		},
		{
			CaseName: "Email inexistant",
			RequestData: map[string]interface{}{
				"email":    "notfound@example.com",
				"password": "password123",
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "Email ou mot de passe incorrect", // À adapter
			SetupUser: func() (string, string, func()) {
				cleanup := func() {}
				return "", "", cleanup
			},
		},
		{
			CaseName: "JSON invalide (champs manquants)",
			RequestData: map[string]interface{}{
				"email": "json.invalide@example.com",
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    "Données invalides", // À adapter
			SetupUser: func() (string, string, func()) {
				cleanup := func() {}
				return "", "", cleanup
			},
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			email, _, cleanup := testCase.SetupUser()
			defer cleanup()

			// Adapter l'email dans la requête si besoin
			if testCase.RequestData["email"] == "login.succes@example.com" || testCase.RequestData["email"] == "login.badpass@example.com" {
				testCase.RequestData["email"] = email
			}

			jsonData, _ := json.Marshal(testCase.RequestData)
			req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if testCase.ExpectedError != "" {
				require.Contains(t, response["error"], testCase.ExpectedError)
			}
			if testCase.ExpectedMessage != "" {
				require.Contains(t, response["message"], testCase.ExpectedMessage)
			}
		})
	}
}

// TestRefreshToken teste la route POST /auth/refresh
func TestRefreshToken(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		RequestData      map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		SetupUser        func() (refreshToken string, cleanup func())
	}{
		{
			CaseName: "Refresh succès (refresh token valide)",
			RequestData: map[string]interface{}{
				"refresh_token": "valid-refresh-token",
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Token rafraîchi avec succès", // À adapter
			ExpectedError:    "",
			SetupUser: func() (string, func()) {
				user, _, err := testutils.CreateAuthenticatedUser(11, "Refresh", "User", testutils.GenerateUniqueEmail("refresh.success"))
				cleanup := func() { _ = testutils.PurgeTestData(user.Email) }
				if err != nil {
					return "", cleanup
				}
				// Récupérer le refresh_token de la session créée
				refreshToken, err := testutils.GetRefreshTokenForUser(user.UserID)
				if err != nil {
					return "", cleanup
				}
				return refreshToken, cleanup
			},
		},
		{
			CaseName: "Refresh token invalide",
			RequestData: map[string]interface{}{
				"refresh_token": "invalid-refresh-token",
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "Session invalide", // À adapter
			SetupUser: func() (string, func()) {
				cleanup := func() {}
				return "invalid-refresh-token", cleanup
			},
		},
		{
			CaseName: "Refresh token expiré",
			RequestData: map[string]interface{}{
				"refresh_token": "expired-refresh-token",
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "session expirée", // À adapter
			SetupUser: func() (string, func()) {
				user, _, err := testutils.CreateAuthenticatedUser(12, "Expired", "User", testutils.GenerateUniqueEmail("refresh.expired"))
				cleanup := func() { _ = testutils.PurgeTestData(user.Email) }
				if err != nil {
					return "", cleanup
				}
				// Expirer le refresh_token en base
				refreshToken, err := testutils.GetRefreshTokenForUser(user.UserID)
				if err != nil {
					return "", cleanup
				}
				testutils.ExpireRefreshToken(refreshToken)
				return refreshToken, cleanup
			},
		},
		{
			CaseName:         "JSON invalide (champs manquants)",
			RequestData:      map[string]interface{}{},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    "Données invalides", // À adapter
			SetupUser: func() (string, func()) {
				cleanup := func() {}
				return "", cleanup
			},
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			refreshToken, cleanup := testCase.SetupUser()
			defer cleanup()

			if testCase.RequestData["refresh_token"] == "valid-refresh-token" || testCase.RequestData["refresh_token"] == "expired-refresh-token" {
				testCase.RequestData["refresh_token"] = refreshToken
			}

			jsonData, _ := json.Marshal(testCase.RequestData)
			req, err := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if testCase.ExpectedError != "" {
				require.Contains(t, response["error"], testCase.ExpectedError)
			}
			if testCase.ExpectedMessage != "" {
				require.Contains(t, response["message"], testCase.ExpectedMessage)
			}
		})
	}
}
