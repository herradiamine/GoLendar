package session_test

import (
	"encoding/json"
	"go-averroes/internal/middleware"
	"go-averroes/internal/session"
	"go-averroes/internal/user"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-averroes/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func createTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// ===== ROUTES D'AUTHENTIFICATION (protégées) =====
	authProtectedGroup := router.Group("/auth")
	authProtectedGroup.Use(middleware.AuthMiddleware())
	{
		authProtectedGroup.POST("/logout", func(c *gin.Context) { session.Session.Logout(c) })
		authProtectedGroup.GET("/me", func(c *gin.Context) { user.User.GetAuthMe(c) })
		authProtectedGroup.GET("/sessions", func(c *gin.Context) { session.Session.GetUserSessions(c) })
		authProtectedGroup.DELETE("/sessions/:session_id", func(c *gin.Context) { session.Session.DeleteSession(c) })
	}

	return router
}

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

	router := createTestRouter()
	gin.SetMode(gin.TestMode)

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

	router := createTestRouter()
	gin.SetMode(gin.TestMode)

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

	router := createTestRouter()
	gin.SetMode(gin.TestMode)

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
