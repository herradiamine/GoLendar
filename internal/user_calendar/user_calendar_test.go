package user_calendar_test

import (
	"encoding/json"
	"fmt"
	"go-averroes/internal/common"
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

// TestGetUserCalendarRoute teste la route GET /user-calendar/:user_id/:calendar_id avec plusieurs cas
func TestGetUserCalendarRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie d'une liaison user-calendar existante",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible avec calendrier et événements
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Construire l'URL avec les vrais IDs
				url := fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, targetUser.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": targetUser.Calendar.CalendarID,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération avec utilisateur inexistant",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
					"url":   "/user-calendar/99999/1",
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de récupération avec calendrier inexistant",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"url":        fmt.Sprintf("/user-calendar/%d/99999", targetUser.User.UserID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération avec liaison user-calendar inexistante",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible sans calendrier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer un calendrier séparé sans liaison avec l'utilisateur
				result, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				calendarID, _ := result.LastInsertId()

				// Ne pas créer la liaison user-calendar pour tester le cas d'erreur

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": calendarID,
					"userID":     targetUser.User.UserID,
					"url":        fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, calendarID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération sans authentification",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Aucune préparation nécessaire, on teste l'accès non authentifié
				return map[string]interface{}{
					"url": "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération sans droits admin",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur normal (non admin) pour tester l'accès refusé
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"url":  "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// On traite les cas de test un par un.
			var req *http.Request
			var err error

			// Préparer la requête HTTP
			if admin, exists := setupData["admin"]; exists {
				// Cas avec authentification admin
				adminUser := admin.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				// Cas avec authentification utilisateur normal
				normalUser := user.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			} else {
				// Cas sans authentification
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
			}

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			// Vérifier la structure de la réponse pour le cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				require.NotNil(t, response.Data)

				// Vérifier que les données contiennent bien une liaison user-calendar
				userCalendarData, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				require.Contains(t, userCalendarData, "user_calendar_id")
				require.Contains(t, userCalendarData, "user_id")
				require.Contains(t, userCalendarData, "calendar_id")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetUserCalendarListRoute teste la route GET /user-calendar/:user_id avec plusieurs cas
func TestGetUserCalendarListRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie des calendriers d'un utilisateur avec calendriers",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible avec calendrier et événements
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Construire l'URL avec l'ID de l'utilisateur cible
				url := fmt.Sprintf("/user-calendar/%d", targetUser.User.UserID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie des calendriers d'un utilisateur sans calendriers",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible sans calendrier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Construire l'URL avec l'ID de l'utilisateur cible
				url := fmt.Sprintf("/user-calendar/%d", targetUser.User.UserID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération avec utilisateur inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
					"url":   "/user-calendar/99999",
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de récupération sans authentification",
			SetupData: func() map[string]interface{} {
				// Aucune préparation nécessaire, on teste l'accès non authentifié
				return map[string]interface{}{
					"url": "/user-calendar/1",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération sans droits admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal (non admin) pour tester l'accès refusé
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"url":  "/user-calendar/1",
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de récupération quand un admin tente d'accéder aux calendriers d'un autre admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour se connecter (sans calendrier)
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un autre utilisateur admin avec calendrier
				otherAdmin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec l'ID de l'autre admin
				url := fmt.Sprintf("/user-calendar/%d/%d", otherAdmin.User.UserID, otherAdmin.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"otherAdmin": otherAdmin,
					"userID":     otherAdmin.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// On traite les cas de test un par un.
			var req *http.Request
			var err error

			// Préparer la requête HTTP
			if admin, exists := setupData["admin"]; exists {
				// Cas avec authentification admin
				adminUser := admin.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				// Cas avec authentification utilisateur normal
				normalUser := user.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			} else {
				// Cas sans authentification
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
			}

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			// Vérifier la structure de la réponse pour le cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)

				// Vérifier que les données contiennent bien une liste de liaisons user-calendar
				userCalendarsData, _ := response.Data.([]interface{})

				// Si l'utilisateur a des calendriers, vérifier la structure
				if len(userCalendarsData) > 0 {
					firstCalendar, ok := userCalendarsData[0].(map[string]interface{})
					require.True(t, ok)
					require.Contains(t, firstCalendar, "user_calendar_id")
					require.Contains(t, firstCalendar, "user_id")
					require.Contains(t, firstCalendar, "calendar_id")
					require.Contains(t, firstCalendar, "title")
					require.Contains(t, firstCalendar, "description")
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetUserCalendarSpecificRoute teste la route GET /user-calendar/:user_id/:calendar_id avec plusieurs cas
func TestGetUserCalendarSpecificRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie d'une liaison user-calendar existante",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible avec calendrier et événements
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Construire l'URL avec les vrais IDs
				url := fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, targetUser.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": targetUser.Calendar.CalendarID,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération avec utilisateur inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
					"url":   "/user-calendar/99999/1",
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de récupération avec calendrier inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"url":        fmt.Sprintf("/user-calendar/%d/99999", targetUser.User.UserID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération avec liaison user-calendar inexistante",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible sans calendrier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer un calendrier séparé sans liaison avec l'utilisateur
				result, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				calendarID, _ := result.LastInsertId()

				// Ne pas créer la liaison user-calendar pour tester le cas d'erreur

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": calendarID,
					"userID":     targetUser.User.UserID,
					"url":        fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, calendarID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération sans authentification",
			SetupData: func() map[string]interface{} {
				// Aucune préparation nécessaire, on teste l'accès non authentifié
				return map[string]interface{}{
					"url": "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération sans droits admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal (non admin) pour tester l'accès refusé
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"url":  "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de récupération quand un admin tente d'accéder à la liaison d'un autre admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour se connecter (sans calendrier)
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un autre utilisateur admin avec calendrier
				otherAdmin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec l'ID de l'autre admin et son calendrier
				url := fmt.Sprintf("/user-calendar/%d/%d", otherAdmin.User.UserID, otherAdmin.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"otherAdmin": otherAdmin,
					"userID":     otherAdmin.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// On traite les cas de test un par un.
			var req *http.Request
			var err error

			// Préparer la requête HTTP
			if admin, exists := setupData["admin"]; exists {
				// Cas avec authentification admin
				adminUser := admin.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				// Cas avec authentification utilisateur normal
				normalUser := user.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			} else {
				// Cas sans authentification
				req, err = http.NewRequest("GET", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
			}

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			// Vérifier la structure de la réponse pour le cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				require.NotNil(t, response.Data)

				// Vérifier que les données contiennent bien une liaison user-calendar
				userCalendarData, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				require.Contains(t, userCalendarData, "user_calendar_id")
				require.Contains(t, userCalendarData, "user_id")
				require.Contains(t, userCalendarData, "calendar_id")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestPostUserCalendarRoute teste la route POST /user-calendar/:user_id/:calendar_id avec plusieurs cas
func TestPostUserCalendarRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Création réussie d'une nouvelle liaison user-calendar",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible sans calendrier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer un calendrier séparé sans liaison avec l'utilisateur
				result, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				calendarID, _ := result.LastInsertId()

				// Construire l'URL avec les vrais IDs
				url := fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, calendarID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": calendarID,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de création avec utilisateur inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
					"url":   "/user-calendar/99999/1",
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de création avec calendrier inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"url":        fmt.Sprintf("/user-calendar/%d/99999", targetUser.User.UserID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de création avec liaison user-calendar déjà existante",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible avec calendrier (liaison déjà existante)
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec les vrais IDs (liaison déjà existante)
				url := fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, targetUser.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": targetUser.Calendar.CalendarID,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserCalendarConflict,
		},
		{
			CaseName: "Échec de création sans authentification",
			SetupData: func() map[string]interface{} {
				// Aucune préparation nécessaire, on teste l'accès non authentifié
				return map[string]interface{}{
					"url": "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de création sans droits admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal (non admin) pour tester l'accès refusé
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"url":  "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de création quand un admin tente de créer une liaison pour un autre admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour se connecter (sans calendrier)
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un autre utilisateur admin avec calendrier
				otherAdmin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec l'ID de l'autre admin et son calendrier
				url := fmt.Sprintf("/user-calendar/%d/%d", otherAdmin.User.UserID, otherAdmin.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"otherAdmin": otherAdmin,
					"userID":     otherAdmin.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// On traite les cas de test un par un.
			var req *http.Request
			var err error

			// Préparer la requête HTTP
			if admin, exists := setupData["admin"]; exists {
				// Cas avec authentification admin
				adminUser := admin.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("POST", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				// Cas avec authentification utilisateur normal
				normalUser := user.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("POST", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			} else {
				// Cas sans authentification
				req, err = http.NewRequest("POST", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
			}

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			// Vérifier la structure de la réponse pour le cas de succès
			if testCase.ExpectedHttpCode == http.StatusCreated {
				require.True(t, response.Success)
				require.NotNil(t, response.Data)

				// Vérifier que les données contiennent bien l'ID de la liaison créée
				userCalendarData, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				require.Contains(t, userCalendarData, "user_calendar_id")
				require.NotNil(t, userCalendarData["user_calendar_id"])
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestPutUserCalendarRoute teste la route PUT /user-calendar/:user_id/:calendar_id avec plusieurs cas
func TestPutUserCalendarRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Mise à jour réussie d'une liaison user-calendar existante",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible avec calendrier (liaison existante)
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec les vrais IDs
				url := fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, targetUser.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": targetUser.Calendar.CalendarID,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de mise à jour avec utilisateur inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
					"url":   "/user-calendar/99999/1",
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de mise à jour avec calendrier inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"url":        fmt.Sprintf("/user-calendar/%d/99999", targetUser.User.UserID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de mise à jour avec liaison user-calendar inexistante",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible sans calendrier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer un calendrier séparé sans liaison avec l'utilisateur
				result, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				calendarID, _ := result.LastInsertId()

				// Ne pas créer la liaison user-calendar pour tester le cas d'erreur

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": calendarID,
					"userID":     targetUser.User.UserID,
					"url":        fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, calendarID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserCalendarNotFound,
		},
		{
			CaseName: "Échec de mise à jour sans authentification",
			SetupData: func() map[string]interface{} {
				// Aucune préparation nécessaire, on teste l'accès non authentifié
				return map[string]interface{}{
					"url": "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de mise à jour sans droits admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal (non admin) pour tester l'accès refusé
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"url":  "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de mise à jour quand un admin tente de modifier une liaison d'un autre admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour se connecter (sans calendrier)
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un autre utilisateur admin avec calendrier
				otherAdmin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec l'ID de l'autre admin et son calendrier
				url := fmt.Sprintf("/user-calendar/%d/%d", otherAdmin.User.UserID, otherAdmin.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"otherAdmin": otherAdmin,
					"userID":     otherAdmin.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// On traite les cas de test un par un.
			var req *http.Request
			var err error

			// Préparer la requête HTTP
			if admin, exists := setupData["admin"]; exists {
				// Cas avec authentification admin
				adminUser := admin.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("PUT", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				// Cas avec authentification utilisateur normal
				normalUser := user.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("PUT", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			} else {
				// Cas sans authentification
				req, err = http.NewRequest("PUT", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
			}

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			// Vérifier la structure de la réponse pour le cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				// La fonction Update ne retourne pas de données, seulement un message de succès
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestDeleteUserCalendarRoute teste la route DELETE /user-calendar/:user_id/:calendar_id avec plusieurs cas
func TestDeleteUserCalendarRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Suppression réussie d'une liaison user-calendar existante",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible avec calendrier (liaison existante)
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec les vrais IDs
				url := fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, targetUser.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": targetUser.Calendar.CalendarID,
					"userID":     targetUser.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression avec utilisateur inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
					"url":   "/user-calendar/99999/1",
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de suppression avec calendrier inexistant",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"url":        fmt.Sprintf("/user-calendar/%d/99999", targetUser.User.UserID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de suppression avec liaison user-calendar inexistante",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour accéder à la route
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible sans calendrier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer un calendrier séparé sans liaison avec l'utilisateur
				result, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				calendarID, _ := result.LastInsertId()

				// Ne pas créer la liaison user-calendar pour tester le cas d'erreur

				return map[string]interface{}{
					"admin":      admin,
					"targetUser": targetUser,
					"calendarID": calendarID,
					"userID":     targetUser.User.UserID,
					"url":        fmt.Sprintf("/user-calendar/%d/%d", targetUser.User.UserID, calendarID),
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserCalendarNotFound,
		},
		{
			CaseName: "Échec de suppression sans authentification",
			SetupData: func() map[string]interface{} {
				// Aucune préparation nécessaire, on teste l'accès non authentifié
				return map[string]interface{}{
					"url": "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de suppression sans droits admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal (non admin) pour tester l'accès refusé
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"url":  "/user-calendar/1/1",
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de suppression quand un admin tente de supprimer une liaison d'un autre admin",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin pour se connecter (sans calendrier)
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un autre utilisateur admin avec calendrier
				otherAdmin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, false)
				require.NoError(t, err)

				// Construire l'URL avec l'ID de l'autre admin et son calendrier
				url := fmt.Sprintf("/user-calendar/%d/%d", otherAdmin.User.UserID, otherAdmin.Calendar.CalendarID)

				return map[string]interface{}{
					"admin":      admin,
					"otherAdmin": otherAdmin,
					"userID":     otherAdmin.User.UserID,
					"url":        url,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// On traite les cas de test un par un.
			var req *http.Request
			var err error

			// Préparer la requête HTTP
			if admin, exists := setupData["admin"]; exists {
				// Cas avec authentification admin
				adminUser := admin.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("DELETE", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+adminUser.SessionToken)
			} else if user, exists := setupData["user"]; exists {
				// Cas avec authentification utilisateur normal
				normalUser := user.(*testutils.AuthenticatedUser)
				req, err = http.NewRequest("DELETE", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+normalUser.SessionToken)
			} else {
				// Cas sans authentification
				req, err = http.NewRequest("DELETE", testServer.URL+setupData["url"].(string), nil)
				require.NoError(t, err)
			}

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error)
			}

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message)
			}

			// Vérifier la structure de la réponse pour le cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success)
				// La fonction Delete ne retourne pas de données, seulement un message de succès
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}
