package user

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
	"go-averroes/internal/middleware"
	"go-averroes/internal/session"
	"go-averroes/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// createTestRouter crée un routeur Gin avec les routes utilisateur configurées comme en production
func createTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// ===== ROUTES D'AUTHENTIFICATION (protégées) =====
	authProtectedGroup := router.Group("/auth")
	authProtectedGroup.Use(middleware.AuthMiddleware())
	{
		authProtectedGroup.POST("/logout", func(c *gin.Context) { session.Session.Logout(c) })
		authProtectedGroup.GET("/me", func(c *gin.Context) { User.GetUserWithRoles(c) })
		authProtectedGroup.GET("/sessions", func(c *gin.Context) { session.Session.GetUserSessions(c) })
		authProtectedGroup.DELETE("/sessions/:session_id", func(c *gin.Context) { session.Session.DeleteSession(c) })
	}

	// ===== ROUTES DE GESTION DES UTILISATEURS =====
	userGroup := router.Group("/user")
	{
		// Création d'utilisateur (public - inscription)
		userGroup.POST("", func(c *gin.Context) { User.Add(c) })

		// Routes protégées par authentification
		userProtectedGroup := userGroup.Group("")
		userProtectedGroup.Use(middleware.AuthMiddleware())
		{
			// L'utilisateur peut accéder à ses propres données
			userProtectedGroup.GET("/me", func(c *gin.Context) { User.Get(c) })
			userProtectedGroup.PUT("/me", func(c *gin.Context) { User.Update(c) })
			userProtectedGroup.DELETE("/me", func(c *gin.Context) { User.Delete(c) })
		}

		// Routes admin pour gérer tous les utilisateurs
		userAdminGroup := userGroup.Group("")
		userAdminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			userAdminGroup.GET("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.Get(c) })
			userAdminGroup.PUT("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.Update(c) })
			userAdminGroup.DELETE("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.Delete(c) })
			userAdminGroup.GET("/:user_id/with-roles", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { User.GetUserWithRoles(c) })
		}
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

// TestUserAdd teste la fonction d'ajout d'utilisateur avec plusieurs cas
func TestUserAdd(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		RequestData      common.CreateUserRequest
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Création d'utilisateur valide",
			RequestData: common.CreateUserRequest{
				Lastname:  "Dupont",
				Firstname: "Jean",
				Email:     "jean.dupont@example.com",
				Password:  "password123",
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Email invalide",
			RequestData: common.CreateUserRequest{
				Lastname:  "Martin",
				Firstname: "Pierre",
				Email:     "email-invalide",
				Password:  "password123",
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Mot de passe trop court",
			RequestData: common.CreateUserRequest{
				Lastname:  "Bernard",
				Firstname: "Marie",
				Email:     "marie.bernard@example.com",
				Password:  "123",
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Données manquantes",
			RequestData: common.CreateUserRequest{
				Lastname:  "",
				Firstname: "",
				Email:     "",
				Password:  "",
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
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/user", User.Add)

			// On prépare les données utiles au traitement de ce cas.
			jsonData, err := json.Marshal(testCase.RequestData)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			if testCase.ExpectedMessage != "" {
				require.Contains(t, response.Message, testCase.ExpectedMessage)
			}

			// On purge les données après avoir traité le cas.
			// Ne pas passer au cas suivant si la purge échoue
			if err := testutils.PurgeTestData(testCase.RequestData.Email); err != nil {
				t.Fatalf("Échec de la purge des données de test: %v", err)
			}
		})
	}
}

// TestUserGet teste la fonction de récupération d'utilisateur
func TestUserGet(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié - route /user/me",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - route /user/me",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Utilisateur avec token invalide - route /user/me",
			SetupAuth: func() (string, string, error) {
				return "Bearer invalid-token", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			router := createTestRouter()

			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/user/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserUpdate teste la fonction de mise à jour d'utilisateur
func TestUserUpdate(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		RequestData      common.UpdateUserRequest
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Mise à jour nom et prénom - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Lastname:  common.StringPtr("Martin"),
				Firstname: common.StringPtr("Pierre"),
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour email valide - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("marie.bernard")
				user, token, err := testutils.CreateAuthenticatedUser(2, "Bernard", "Marie", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Email: common.StringPtr(testutils.GenerateUniqueEmail("nouveau.email")),
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Email invalide - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("sophie.petit")
				user, token, err := testutils.CreateAuthenticatedUser(3, "Petit", "Sophie", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Email: common.StringPtr("email-invalide"),
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEmailFormat,
		},
		{
			CaseName: "Mot de passe trop court - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("thomas.roux")
				user, token, err := testutils.CreateAuthenticatedUser(4, "Roux", "Thomas", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Password: common.StringPtr("123"),
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrPasswordTooShort,
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			RequestData: common.UpdateUserRequest{
				Lastname: common.StringPtr("Nouveau"),
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			router := createTestRouter()

			// On prépare les données utiles au traitement de ce cas.
			jsonData, err := json.Marshal(testCase.RequestData)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			if testCase.ExpectedMessage != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Message, testCase.ExpectedMessage)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserDelete teste la fonction de suppression d'utilisateur
func TestUserDelete(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Suppression d'utilisateur valide - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserDelete,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			router := createTestRouter()

			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("DELETE", "/user/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			if testCase.ExpectedMessage != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Message, testCase.ExpectedMessage)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserGetWithRoles teste la fonction de récupération d'utilisateur avec rôles
func TestUserGetWithRoles(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur avec rôles dans le contexte - route /auth/me",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - route /auth/me",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			router := createTestRouter()

			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserGetAuthMe teste la fonction GetAuthMe (alias de GetUserWithRoles)
func TestUserGetAuthMe(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié avec rôles - route /auth/me",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - route /auth/me",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			router := createTestRouter()

			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserAddErrors teste les cas d'erreur de la fonction Add
func TestUserAddErrors(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() (common.CreateUserRequest, string, error) // Retourne (request, emailToPurge, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Email déjà utilisé",
			SetupData: func() (common.CreateUserRequest, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				// Créer un premier utilisateur
				_, err := testutils.CreateUserWithPassword("Dupont", "Jean", email, "password123")
				if err != nil {
					return common.CreateUserRequest{}, "", err
				}
				// Essayer de créer un deuxième utilisateur avec le même email
				req := common.CreateUserRequest{
					Lastname:  "Martin",
					Firstname: "Pierre",
					Email:     email,
					Password:  "password456",
				}
				return req, email, nil
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			router := createTestRouter()

			// On prépare les données utiles au traitement de ce cas.
			req, emailToPurge, err := testCase.SetupData()
			require.NoError(t, err)

			jsonData, err := json.Marshal(req)
			require.NoError(t, err)

			httpReq, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, httpReq)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if emailToPurge != "" {
				if err := testutils.PurgeTestData(emailToPurge); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserUpdateErrors teste les cas d'erreur de la fonction Update
func TestUserUpdateErrors(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, string, error) // Retourne (token, userEmail, emailToPurge, error)
		RequestData      common.UpdateUserRequest
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Email déjà utilisé par un autre utilisateur",
			SetupAuth: func() (string, string, string, error) {
				// Créer un premier utilisateur
				email1 := testutils.GenerateUniqueEmail("jean.dupont")
				user1, _, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email1)
				if err != nil {
					return "", "", "", err
				}

				// Créer un deuxième utilisateur
				email2 := testutils.GenerateUniqueEmail("marie.martin")
				user2, token, err := testutils.CreateAuthenticatedUser(2, "Martin", "Marie", email2)
				if err != nil {
					return "", "", "", err
				}

				// Essayer de mettre à jour le deuxième utilisateur avec l'email du premier
				return "Bearer " + token, user2.Email, user1.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Email: common.StringPtr(""), // Sera défini dynamiquement
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			router := createTestRouter()

			// On prépare les données utiles au traitement de ce cas.
			token, userEmail, emailToPurge, err := testCase.SetupAuth()
			require.NoError(t, err)

			// Si c'est le cas d'email déjà utilisé, on utilise l'email du premier utilisateur
			if testCase.RequestData.Email != nil && *testCase.RequestData.Email == "" {
				testCase.RequestData.Email = common.StringPtr(emailToPurge)
			}

			jsonData, err := json.Marshal(testCase.RequestData)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Configurer l'authentification si nécessaire
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
			if emailToPurge != "" {
				if err := testutils.PurgeTestData(emailToPurge); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserAdminRoutes teste les routes admin pour la gestion des utilisateurs
func TestUserAdminRoutes(t *testing.T) {
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (token string, userEmail string, url string, err error)
		Method           string
		RequestData      interface{}
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Admin récupère un utilisateur par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID)
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method:           "GET",
			RequestData:      nil,
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin met à jour un utilisateur par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID)
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method: "PUT",
			RequestData: common.UpdateUserRequest{
				Lastname: common.StringPtr("Updated"),
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin supprime un utilisateur par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID)
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method:           "DELETE",
			RequestData:      nil,
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin récupère un utilisateur avec rôles par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID) + "/with-roles"
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method:           "GET",
			RequestData:      nil,
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			router := createTestRouter()
			token, userEmail, url, err := testCase.SetupAuth()
			require.NoError(t, err)
			var req *http.Request
			if testCase.RequestData != nil {
				jsonData, err := json.Marshal(testCase.RequestData)
				require.NoError(t, err)
				req, err = http.NewRequest(testCase.Method, url, bytes.NewBuffer(jsonData))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(testCase.Method, url, nil)
				require.NoError(t, err)
			}
			if token != "" {
				req.Header.Set("Authorization", token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}
