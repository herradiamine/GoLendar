package middleware_test

import (
	"encoding/json"
	"fmt"
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var testGinRouter *gin.Engine // Routeur de test global
var testServer *httptest.Server
var testClient *http.Client

// TestMain configure l'environnement de test global
func TestMain(m *testing.M) {
	if err := testutils.SetupTestEnvironment(); err != nil {
		panic("Impossible d'initialiser l'environnement de test: " + err.Error())
	}
	testGinRouter = testutils.CreateTestRouter()
	testServer = httptest.NewServer(testGinRouter)
	testClient = testServer.Client()
	code := m.Run()
	if err := testutils.TeardownTestEnvironment(); err != nil {
		panic("Impossible de nettoyer l'environnement de test: " + err.Error())
	}
	testServer.Close()
	os.Exit(code)
}

// TestUserExistsMiddleware teste le middleware UserExistsMiddleware
func TestUserExistsMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		URL              string
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur existant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur pour le test
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			URL:              "/test-user/1",
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur inexistant",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			URL:              "/test-user/99999",
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "ID utilisateur invalide",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			URL:              "/test-user/invalid",
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedError:    common.ErrInvalidUserID,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			testCase.SetupData()

			// Créer un routeur de test avec le middleware
			router := gin.New()
			router.GET("/test-user/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) {
				user, exists := c.Get("user")
				if exists {
					c.JSON(http.StatusOK, gin.H{"user": user})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no user"})
				}
			})

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+testCase.URL, nil)
			require.NoError(t, err)

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestCalendarExistsMiddleware teste le middleware CalendarExistsMiddleware
func TestCalendarExistsMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		URL              string
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Calendrier existant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur avec calendrier pour le test
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			URL:              "/test-calendar/1",
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Calendrier inexistant",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			URL:              "/test-calendar/99999",
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "ID calendrier invalide",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			URL:              "/test-calendar/invalid",
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedError:    common.ErrInvalidCalendarID,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			testCase.SetupData()

			// Créer un routeur de test avec le middleware
			router := gin.New()
			router.GET("/test-calendar/:calendar_id", middleware.CalendarExistsMiddleware("calendar_id"), func(c *gin.Context) {
				calendar, exists := c.Get("calendar")
				if exists {
					c.JSON(http.StatusOK, gin.H{"calendar": calendar})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no calendar"})
				}
			})

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+testCase.URL, nil)
			require.NoError(t, err)

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestAuthMiddleware teste le middleware AuthMiddleware
func TestAuthMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		AuthHeader       string
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Authentification réussie",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur authentifié
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			AuthHeader:       "Bearer valid-token", // Sera remplacé par le vrai token
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Header Authorization manquant",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			AuthHeader:       "",
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Token invalide",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			AuthHeader:       "Bearer invalid-token",
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Format de token invalide",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			AuthHeader:       "InvalidFormat token",
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			setupData := testCase.SetupData()

			// Créer un routeur de test avec le middleware
			router := gin.New()
			router.GET("/test-auth", middleware.AuthMiddleware(), func(c *gin.Context) {
				user, exists := c.Get("auth_user")
				if exists {
					c.JSON(http.StatusOK, gin.H{"user": user})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no user"})
				}
			})

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+"/test-auth", nil)
			require.NoError(t, err)

			// Ajouter le header d'authentification
			if user, exists := setupData["user"]; exists {
				// Utiliser le vrai token de session
				userData := user.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+userData.SessionToken)
			} else if testCase.AuthHeader != "" {
				// Utiliser le header fourni dans le test
				req.Header.Set("Authorization", testCase.AuthHeader)
			}

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestAdminMiddleware teste le middleware AdminMiddleware
func TestAdminMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			setupData := testCase.SetupData()

			// Créer un routeur de test avec les middlewares
			router := gin.New()
			router.GET("/test-admin", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
			})

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+"/test-admin", nil)
			require.NoError(t, err)

			// Ajouter le header d'authentification
			if admin, exists := setupData["admin"]; exists {
				adminUser := admin.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				normalUser := user.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			}

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestUserCanAccessCalendarMiddleware teste le middleware UserCanAccessCalendarMiddleware
func TestUserCanAccessCalendarMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur avec accès au calendrier",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur avec calendrier (accès automatique)
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur sans accès au calendrier",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur sans calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer un calendrier séparé sans liaison avec l'utilisateur
				result, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				calendarID, _ := result.LastInsertId()

				return map[string]interface{}{
					"user":       user,
					"calendarID": calendarID,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			setupData := testCase.SetupData()

			// Créer un routeur de test avec les middlewares
			router := gin.New()
			router.GET("/test-calendar-access/:calendar_id",
				middleware.AuthMiddleware(),
				middleware.CalendarExistsMiddleware("calendar_id"),
				middleware.UserCanAccessCalendarMiddleware(),
				func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "calendar access granted"})
				},
			)

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			calendarID := "1" // Par défaut
			if id, exists := setupData["calendarID"]; exists {
				calendarID = fmt.Sprintf("%d", id)
			}

			req, err := http.NewRequest("GET", server.URL+"/test-calendar-access/"+calendarID, nil)
			require.NoError(t, err)

			// Ajouter le header d'authentification
			if user, exists := setupData["user"]; exists {
				userData := user.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+userData.SessionToken)
			}

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestOptionalAuthMiddleware teste le middleware OptionalAuthMiddleware
func TestOptionalAuthMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		AuthHeader       string
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Avec authentification valide",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur authentifié
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			AuthHeader:       "Bearer valid-token",
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Sans authentification",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			AuthHeader:       "",
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Avec token invalide",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			AuthHeader:       "Bearer invalid-token",
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			testCase.SetupData()

			// Créer un routeur de test avec le middleware
			router := gin.New()
			router.GET("/test-optional-auth", middleware.OptionalAuthMiddleware(), func(c *gin.Context) {
				user, exists := c.Get("auth_user")
				if exists {
					c.JSON(http.StatusOK, gin.H{"user": user})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no user"})
				}
			})

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+"/test-optional-auth", nil)
			require.NoError(t, err)

			// Ajouter le header d'authentification si fourni
			if testCase.AuthHeader != "" {
				req.Header.Set("Authorization", testCase.AuthHeader)
			}

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestEventExistsMiddleware teste le middleware EventExistsMiddleware
func TestEventExistsMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		URL              string
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Événement existant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur avec calendrier et événement
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			URL:              "/test-event/1",
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Événement inexistant",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			URL:              "/test-event/99999",
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "ID événement invalide",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			URL:              "/test-event/invalid",
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedError:    common.ErrInvalidEventID,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			testCase.SetupData()

			// Créer un routeur de test avec le middleware
			router := gin.New()
			router.GET("/test-event/:event_id", middleware.EventExistsMiddleware("event_id"), func(c *gin.Context) {
				event, exists := c.Get("event")
				if exists {
					c.JSON(http.StatusOK, gin.H{"event": event})
				} else {
					c.JSON(http.StatusOK, gin.H{"message": "no event"})
				}
			})

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+testCase.URL, nil)
			require.NoError(t, err)

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestRoleMiddleware teste le middleware RoleMiddleware
func TestRoleMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		RequiredRole     string
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur avec le rôle requis",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin (qui a le rôle "admin")
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
				}
			},
			RequiredRole:     "admin",
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur sans le rôle requis",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal (sans rôle admin)
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			RequiredRole:     "admin",
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Utilisateur avec un rôle différent",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			RequiredRole:     "moderator",
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			setupData := testCase.SetupData()

			// Créer un routeur de test avec les middlewares
			router := gin.New()
			router.GET("/test-role",
				middleware.AuthMiddleware(),
				middleware.RoleMiddleware(testCase.RequiredRole),
				func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "role access granted"})
				},
			)

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+"/test-role", nil)
			require.NoError(t, err)

			// Ajouter le header d'authentification
			if admin, exists := setupData["admin"]; exists {
				adminUser := admin.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				normalUser := user.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			}

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestRolesMiddleware teste le middleware RolesMiddleware
func TestRolesMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		RequiredRoles    []string
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur avec un des rôles requis",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin (qui a le rôle "admin")
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
				}
			},
			RequiredRoles:    []string{"admin", "moderator"},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur avec le premier rôle de la liste",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
				}
			},
			RequiredRoles:    []string{"admin"},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur sans aucun des rôles requis",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal (sans rôle admin ou moderator)
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			RequiredRoles:    []string{"moderator", "editor"},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Utilisateur avec un rôle différent",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			RequiredRoles:    []string{"moderator"},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			setupData := testCase.SetupData()

			// Créer un routeur de test avec les middlewares
			router := gin.New()
			router.GET("/test-roles",
				middleware.AuthMiddleware(),
				middleware.RolesMiddleware(testCase.RequiredRoles...),
				func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "roles access granted"})
				},
			)

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest("GET", server.URL+"/test-roles", nil)
			require.NoError(t, err)

			// Ajouter le header d'authentification
			if admin, exists := setupData["admin"]; exists {
				adminUser := admin.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				normalUser := user.(*testutils.AuthenticatedUser)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			}

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Si on attend une erreur, vérifier le message
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestLoggingMiddleware teste le middleware LoggingMiddleware
func TestLoggingMiddleware(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		Method           string
		URL              string
		ExpectedHttpCode int
	}{
		{
			CaseName: "Requête GET réussie",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			Method:           "GET",
			URL:              "/test-logging",
			ExpectedHttpCode: http.StatusOK,
		},
		{
			CaseName: "Requête POST réussie",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			Method:           "POST",
			URL:              "/test-logging",
			ExpectedHttpCode: http.StatusOK,
		},
		{
			CaseName: "Requête avec erreur 404",
			SetupData: func() map[string]interface{} {
				return map[string]interface{}{}
			},
			Method:           "GET",
			URL:              "/not-found",
			ExpectedHttpCode: http.StatusNotFound,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données
			testCase.SetupData()

			// Créer un routeur de test avec le middleware de logging
			router := gin.New()
			router.Use(middleware.LoggingMiddleware())

			// Route de test qui fonctionne
			router.GET("/test-logging", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			router.POST("/test-logging", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Créer un serveur de test
			server := httptest.NewServer(router)
			defer server.Close()

			// Faire la requête
			req, err := http.NewRequest(testCase.Method, server.URL+testCase.URL, nil)
			require.NoError(t, err)

			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Le middleware de logging ne modifie pas la réponse, il se contente de logger
			// On vérifie juste que la requête a été traitée correctement
			if testCase.ExpectedHttpCode == http.StatusOK {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, "success", response["message"])
			}

			// Nettoyer les données
			testutils.PurgeAllTestUsers()
		})
	}
}
