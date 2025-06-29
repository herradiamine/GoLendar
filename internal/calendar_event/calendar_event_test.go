package calendar_event_test

import (
	"bytes"
	"encoding/json"
	"go-averroes/internal/common"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

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

// TestGetEventRoute teste la route GET de récupération d'un événement par ID avec plusieurs cas
func TestGetEventRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie d'un événement par son propriétaire",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'un événement partagé par un autre utilisateur",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier et de l'événement
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un utilisateur qui aura accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Partager le calendrier avec l'utilisateur
				_, err = common.DB.Exec(`
					INSERT INTO user_calendar (user_id, calendar_id, created_at) 
					VALUES (?, ?, NOW())
				`, user.User.UserID, 1) // Le premier calendrier créé aura l'ID 1
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec token invalide",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session expirée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec calendar_id inexistant",
			CaseUrl:  "/calendar-event/99999/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/invalid/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de récupération avec event_id inexistant",
			CaseUrl:  "/calendar-event/1/99999", // Sera remplacé par l'ID du calendrier réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de récupération avec event_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/invalid", // Sera remplacé par l'ID du calendrier réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEventID,
		},
		{
			CaseName: "Échec de récupération sans accès au calendrier",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier et de l'événement
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un autre utilisateur sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de récupération d'un événement supprimé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer l'événement (soft delete)
				_, err = common.DB.Exec(`
					UPDATE event 
					SET deleted_at = NOW() 
					WHERE event_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de récupération d'un événement d'un calendrier supprimé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer le calendrier (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// Remplacer les IDs dans l'URL si nécessaire
			url := testCase.CaseUrl
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				// Récupérer l'ID du calendrier et de l'événement de l'utilisateur
				var calendarID, eventID int
				err := common.DB.QueryRow(`
					SELECT c.calendar_id, e.event_id 
					FROM calendar c
					INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
					INNER JOIN calendar_event ce ON c.calendar_id = ce.calendar_id
					INNER JOIN event e ON ce.event_id = e.event_id
					WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL 
					  AND e.deleted_at IS NULL AND ce.deleted_at IS NULL
					ORDER BY c.created_at DESC, e.created_at DESC
					LIMIT 1
				`, user.User.UserID).Scan(&calendarID, &eventID)
				if err == nil {
					url = "/calendar-event/" + strconv.Itoa(calendarID) + "/" + strconv.Itoa(eventID)
				}
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header d'authentification si disponible
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+user.SessionToken)
			} else if sessionToken, ok := setupData["sessionToken"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+sessionToken)
			} else if authHeader, ok := setupData["authHeader"].(string); ok {
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}
			}

			// On traite les cas de test un par un.
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
			}

			// Vérifier le message d'erreur
			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError, "Message d'erreur incorrect")
			}

			// Vérifications spécifiques pour les cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")
				require.NotNil(t, response.Data, "Les données de réponse ne devraient pas être nulles")

				// Vérifier que les données de l'événement sont présentes
				eventData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet événement")
				require.Contains(t, eventData, "event_id", "L'événement devrait avoir un ID")
				require.Contains(t, eventData, "title", "L'événement devrait avoir un titre")
				require.Contains(t, eventData, "start", "L'événement devrait avoir une date de début")
				require.Contains(t, eventData, "duration", "L'événement devrait avoir une durée")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestAddEventRoute teste la route POST de création d'un événement avec plusieurs cas
func TestAddEventRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		RequestData      func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Création réussie d'un événement avec toutes les données",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Réunion équipe",
						"description": "Réunion hebdomadaire de l'équipe de développement",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
						"canceled":    false,
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Création réussie d'un événement avec données minimales",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement simple",
						"start":       time.Now().Add(2 * time.Hour).Format(time.RFC3339),
						"duration":    30,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Création réussie d'un événement annulé",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement annulé",
						"start":       time.Now().Add(3 * time.Hour).Format(time.RFC3339),
						"duration":    45,
						"calendar_id": 1,
						"canceled":    true,
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Création réussie d'un événement partagé par un autre utilisateur",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
				require.NoError(t, err)

				// Créer un utilisateur qui aura accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Partager le calendrier avec l'utilisateur
				_, err = common.DB.Exec(`
					INSERT INTO user_calendar (user_id, calendar_id, created_at) 
					VALUES (?, ?, NOW())
				`, user.User.UserID, 1) // Le premier calendrier créé aura l'ID 1
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
					"requestBody": map[string]interface{}{
						"title":       "Événement partagé",
						"start":       time.Now().Add(4 * time.Hour).Format(time.RFC3339),
						"duration":    90,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de création sans header Authorization",
			CaseUrl:  "/calendar-event/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de création avec header Authorization vide",
			CaseUrl:  "/calendar-event/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "",
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de création avec token invalide",
			CaseUrl:  "/calendar-event/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec session expirée",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec session désactivée",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec calendar_id inexistant",
			CaseUrl:  "/calendar-event/99999",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 99999,
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de création avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/invalid",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de création sans accès au calendrier",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
				require.NoError(t, err)

				// Créer un autre utilisateur sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de création avec titre manquant",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec titre vide",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec date de début manquante",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec date de début invalide",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       "date-invalide",
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec durée manquante",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec durée invalide (zéro)",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    0,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec durée invalide (négative)",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    -30,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec données JSON vides",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":        user,
					"requestBody": map[string]interface{}{},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création d'un événement dans un calendrier supprimé",
			CaseUrl:  "/calendar-event/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)

				// Supprimer le calendrier (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement test",
						"start":       time.Now().Add(1 * time.Hour).Format(time.RFC3339),
						"duration":    60,
						"calendar_id": 1,
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			requestData := testCase.RequestData()

			// Remplacer les IDs dans l'URL si nécessaire
			url := testCase.CaseUrl
			if user, ok := requestData["user"].(*testutils.AuthenticatedUser); ok {
				// Récupérer l'ID du calendrier de l'utilisateur
				var calendarID int
				err := common.DB.QueryRow(`
					SELECT c.calendar_id 
					FROM calendar c
					INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
					WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL
					ORDER BY c.created_at DESC
					LIMIT 1
				`, user.User.UserID).Scan(&calendarID)
				if err == nil {
					url = "/calendar-event/" + strconv.Itoa(calendarID)
					// Mettre à jour le calendar_id dans le corps de la requête
					if requestBody, ok := requestData["requestBody"].(map[string]interface{}); ok {
						requestBody["calendar_id"] = calendarID
					}
				}
			}

			// Extraire les données de requête
			requestBody, ok := requestData["requestBody"].(map[string]interface{})
			require.True(t, ok, "Le corps de la requête doit être présent")

			// Préparer la requête JSON
			jsonData, err := json.Marshal(requestBody)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", testServer.URL+url, bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")
			req.Header.Set("Content-Type", "application/json")

			// Ajouter le header d'authentification si disponible
			if user, ok := requestData["user"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+user.SessionToken)
			} else if sessionToken, ok := requestData["sessionToken"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+sessionToken)
			} else if authHeader, ok := requestData["authHeader"].(string); ok {
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}
			}

			// On traite les cas de test un par un.
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
			}

			// Vérifier le message d'erreur
			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError, "Message d'erreur incorrect")
			}

			// Vérifications spécifiques pour les cas de succès
			if testCase.ExpectedHttpCode == http.StatusCreated {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")
				require.NotNil(t, response.Data, "Les données de réponse ne devraient pas être nulles")

				// Vérifier que les données de l'événement créé sont présentes
				eventData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet événement")
				require.Contains(t, eventData, "event_id", "L'événement devrait avoir un ID")
				require.Contains(t, eventData, "calendar_id", "L'événement devrait avoir un ID de calendrier")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestUpdateEventRoute teste la route PUT de mise à jour d'un événement avec plusieurs cas
func TestUpdateEventRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		RequestData      func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Mise à jour réussie du titre",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre d'événement",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour réussie de la description",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"description": "Nouvelle description de l'événement",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour réussie de la date de début",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"start": time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour réussie de la durée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"duration": 90,
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour réussie du statut annulé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"canceled": true,
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour réussie de plusieurs champs",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Événement mis à jour",
						"description": "Description mise à jour",
						"start":       time.Now().Add(3 * time.Hour).Format(time.RFC3339),
						"duration":    120,
						"canceled":    false,
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour réussie par un utilisateur partagé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur propriétaire du calendrier et de l'événement
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un utilisateur qui aura accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Partager le calendrier avec l'utilisateur
				_, err = common.DB.Exec(`
					INSERT INTO user_calendar (user_id, calendar_id, created_at) 
					VALUES (?, ?, NOW())
				`, user.User.UserID, 1) // Le premier calendrier créé aura l'ID 1
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
					"requestBody": map[string]interface{}{
						"title": "Événement modifié par utilisateur partagé",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour réussie avec données JSON vides (aucune modification)",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":        user,
					"requestBody": map[string]interface{}{},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de mise à jour sans header Authorization",
			CaseUrl:  "/calendar-event/1/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier et un événement sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de mise à jour avec header Authorization vide",
			CaseUrl:  "/calendar-event/1/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "",
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de mise à jour avec token invalide",
			CaseUrl:  "/calendar-event/1/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de mise à jour avec session expirée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de mise à jour avec session désactivée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de mise à jour avec calendar_id inexistant",
			CaseUrl:  "/calendar-event/99999/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de mise à jour avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/invalid/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de mise à jour avec event_id inexistant",
			CaseUrl:  "/calendar-event/1/99999", // Sera remplacé par l'ID du calendrier réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de mise à jour avec event_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/invalid", // Sera remplacé par l'ID du calendrier réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de mise à jour sans accès au calendrier",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur propriétaire du calendrier et de l'événement
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un autre utilisateur sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de mise à jour avec date de début invalide",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"start": "date-invalide",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de mise à jour avec durée invalide (zéro)",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"duration": 0,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de mise à jour avec durée invalide (négative)",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"duration": -30,
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de mise à jour d'un événement supprimé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer l'événement (soft delete)
				_, err = common.DB.Exec(`
					UPDATE event 
					SET deleted_at = NOW() 
					WHERE event_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de mise à jour d'un événement d'un calendrier supprimé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer le calendrier (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Nouveau titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			requestData := testCase.RequestData()

			// Remplacer les IDs dans l'URL si nécessaire
			url := testCase.CaseUrl
			if user, ok := requestData["user"].(*testutils.AuthenticatedUser); ok {
				// Récupérer l'ID du calendrier de l'utilisateur
				var calendarID int
				err := common.DB.QueryRow(`
					SELECT c.calendar_id 
					FROM calendar c
					INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
					WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL
					ORDER BY c.created_at DESC
					LIMIT 1
				`, user.User.UserID).Scan(&calendarID)
				if err == nil {
					url = "/calendar-event/" + strconv.Itoa(calendarID) + "/" + strconv.Itoa(calendarID)
				}
			}

			// Extraire les données de requête
			requestBody, ok := requestData["requestBody"].(map[string]interface{})
			require.True(t, ok, "Le corps de la requête doit être présent")

			// Préparer la requête JSON
			jsonData, err := json.Marshal(requestBody)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("PUT", testServer.URL+url, bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")
			req.Header.Set("Content-Type", "application/json")

			// Ajouter le header d'authentification si disponible
			if user, ok := requestData["user"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+user.SessionToken)
			} else if sessionToken, ok := requestData["sessionToken"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+sessionToken)
			} else if authHeader, ok := requestData["authHeader"].(string); ok {
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}
			}

			// On traite les cas de test un par un.
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
			}

			// Vérifier le message d'erreur
			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError, "Message d'erreur incorrect")
			}

			// Vérifications spécifiques pour les cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestDeleteEventRoute teste la route DELETE de suppression d'un événement avec plusieurs cas
func TestDeleteEventRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Suppression réussie d'un événement par son propriétaire",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Suppression réussie d'un événement par un utilisateur partagé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier et de l'événement
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un utilisateur qui aura accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Partager le calendrier avec l'utilisateur
				_, err = common.DB.Exec(`
					INSERT INTO user_calendar (user_id, calendar_id, created_at) 
					VALUES (?, ?, NOW())
				`, user.User.UserID, 1) // Le premier calendrier créé aura l'ID 1
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression sans header Authorization",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de suppression avec header Authorization vide",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de suppression avec token invalide",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec session expirée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec session désactivée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec calendar_id inexistant",
			CaseUrl:  "/calendar-event/99999/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de suppression avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/invalid/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de suppression avec event_id inexistant",
			CaseUrl:  "/calendar-event/1/99999", // Sera remplacé par l'ID du calendrier réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de suppression avec event_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/invalid", // Sera remplacé par l'ID du calendrier réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEventID,
		},
		{
			CaseName: "Échec de suppression sans accès au calendrier",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier et de l'événement
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un autre utilisateur sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de suppression d'un événement déjà supprimé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer l'événement (soft delete)
				_, err = common.DB.Exec(`
					UPDATE event 
					SET deleted_at = NOW() 
					WHERE event_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de suppression d'un événement d'un calendrier supprimé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer le calendrier (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de suppression d'un événement annulé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Marquer l'événement comme annulé
				_, err = common.DB.Exec(`
					UPDATE event 
					SET canceled = TRUE 
					WHERE event_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression avec permissions insuffisantes (utilisateur normal)",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier et de l'événement
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un utilisateur normal sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, false, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec liaison calendar_event supprimée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer la liaison calendar_event (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar_event 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1 AND event_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteEvent,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression avec événement inexistant dans le calendrier",
			CaseUrl:  "/calendar-event/1/99999", // Sera remplacé par l'ID du calendrier réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)

				// Créer un événement qui n'est pas lié au calendrier
				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement isolé", "Description événement isolé", time.Now().Add(2*time.Hour), 60, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrEventNotFound,
		},
		{
			CaseName: "Échec de suppression avec utilisateur admin tentant d'accéder à un autre admin",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un autre utilisateur admin sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de suppression avec format de token invalide",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "InvalidFormat token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec token de session inexistant",
			CaseUrl:  "/calendar-event/1/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier et un événement
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement Test", "Description événement", time.Now().Add(1*time.Hour), 60, false)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, 1)
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer non_existent_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec utilisateur supprimé",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer l'utilisateur (soft delete)
				_, err = common.DB.Exec(`
					UPDATE user 
					SET deleted_at = NOW() 
					WHERE user_id = ?
				`, user.User.UserID)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec liaison user_calendar supprimée",
			CaseUrl:  "/calendar-event/1/1", // Sera remplacé par les IDs réels
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer la liaison user_calendar (soft delete)
				_, err = common.DB.Exec(`
					UPDATE user_calendar 
					SET deleted_at = NOW() 
					WHERE user_id = ? AND calendar_id = 1
				`, user.User.UserID)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// Remplacer les IDs dans l'URL si nécessaire
			url := testCase.CaseUrl
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				// Récupérer l'ID du calendrier et de l'événement de l'utilisateur
				var calendarID, eventID int
				err := common.DB.QueryRow(`
					SELECT c.calendar_id, e.event_id 
					FROM calendar c
					INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
					INNER JOIN calendar_event ce ON c.calendar_id = ce.calendar_id
					INNER JOIN event e ON ce.event_id = e.event_id
					WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL 
					  AND e.deleted_at IS NULL AND ce.deleted_at IS NULL
					ORDER BY c.created_at DESC, e.created_at DESC
					LIMIT 1
				`, user.User.UserID).Scan(&calendarID, &eventID)
				if err == nil {
					url = "/calendar-event/" + strconv.Itoa(calendarID) + "/" + strconv.Itoa(eventID)
				}
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("DELETE", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header d'authentification si disponible
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+user.SessionToken)
			} else if sessionToken, ok := setupData["sessionToken"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+sessionToken)
			} else if authHeader, ok := setupData["authHeader"].(string); ok {
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}
			}

			// On traite les cas de test un par un.
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
			}

			// Vérifier le message d'erreur
			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError, "Message d'erreur incorrect")
			}

			// Vérifications spécifiques pour les cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetEventsByMonthRoute teste la route GET de récupération d'événements par mois avec plusieurs cas
func TestGetEventsByMonthRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie d'événements par mois avec événements",
			CaseUrl:  "/calendar-event/1/month/2024/12", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Créer des événements pour décembre 2024
				result, err := common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement 1 Décembre", "Description événement 1", time.Date(2024, 12, 15, 10, 0, 0, 0, time.UTC), 60, false)
				require.NoError(t, err)
				eventID1, err := result.LastInsertId()
				require.NoError(t, err)

				result, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement 2 Décembre", "Description événement 2", time.Date(2024, 12, 20, 14, 30, 0, 0, time.UTC), 90, false)
				require.NoError(t, err)
				eventID2, err := result.LastInsertId()
				require.NoError(t, err)

				// Associer les événements au calendrier
				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID1)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID2)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'événements par mois sans événements",
			CaseUrl:  "/calendar-event/1/month/2024/11", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'événements par mois par un utilisateur partagé",
			CaseUrl:  "/calendar-event/1/month/2024/12", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un utilisateur qui aura accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Partager le calendrier avec l'utilisateur
				_, err = common.DB.Exec(`
					INSERT INTO user_calendar (user_id, calendar_id, created_at) 
					VALUES (?, ?, NOW())
				`, user.User.UserID, 1) // Le premier calendrier créé aura l'ID 1
				require.NoError(t, err)

				// Créer un événement pour décembre 2024
				result, err := common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement partagé Décembre", "Description événement partagé", time.Date(2024, 12, 25, 16, 0, 0, 0, time.UTC), 120, false)
				require.NoError(t, err)
				eventID, err := result.LastInsertId()
				require.NoError(t, err)

				// Associer l'événement au calendrier
				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/calendar-event/1/month/2024/12",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/calendar-event/1/month/2024/12",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec token invalide",
			CaseUrl:  "/calendar-event/1/month/2024/12",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session expirée",
			CaseUrl:  "/calendar-event/1/month/2024/12", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/calendar-event/1/month/2024/12", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec calendar_id inexistant",
			CaseUrl:  "/calendar-event/99999/month/2024/12",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/invalid/month/2024/12",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de récupération avec année invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/month/invalid/12", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidYear,
		},
		{
			CaseName: "Échec de récupération avec mois invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/month/2024/invalid", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidMonth,
		},
		{
			CaseName: "Échec de récupération avec mois invalide (hors limites)",
			CaseUrl:  "/calendar-event/1/month/2024/13", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidMonth,
		},
		{
			CaseName: "Échec de récupération avec mois invalide (zéro)",
			CaseUrl:  "/calendar-event/1/month/2024/0", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidMonth,
		},
		{
			CaseName: "Échec de récupération sans accès au calendrier",
			CaseUrl:  "/calendar-event/1/month/2024/12", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un autre utilisateur sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de récupération d'événements d'un calendrier supprimé",
			CaseUrl:  "/calendar-event/1/month/2024/12", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer le calendrier (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// Remplacer les IDs dans l'URL si nécessaire
			url := testCase.CaseUrl
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				// Récupérer l'ID du calendrier de l'utilisateur
				var calendarID int
				err := common.DB.QueryRow(`
					SELECT c.calendar_id 
					FROM calendar c
					INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
					WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL
					ORDER BY c.created_at DESC
					LIMIT 1
				`, user.User.UserID).Scan(&calendarID)
				if err == nil {
					// Préserver les paramètres year et month de l'URL originale pour les cas de validation
					if strings.Contains(testCase.CaseName, "invalide") {
						// Pour les cas de validation, garder les paramètres invalides
						urlParts := strings.Split(testCase.CaseUrl, "/")
						if len(urlParts) >= 6 {
							url = "/calendar-event/" + strconv.Itoa(calendarID) + "/month/" + urlParts[4] + "/" + urlParts[5]
						}
					} else {
						// Pour les autres cas, utiliser les paramètres par défaut
						url = "/calendar-event/" + strconv.Itoa(calendarID) + "/month/2024/12"
					}
				}
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header d'authentification si disponible
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+user.SessionToken)
			} else if sessionToken, ok := setupData["sessionToken"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+sessionToken)
			} else if authHeader, ok := setupData["authHeader"].(string); ok {
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}
			}

			// On traite les cas de test un par un.
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
			}

			// Vérifier le message d'erreur
			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError, "Message d'erreur incorrect")
			}

			// Vérifications spécifiques pour les cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")

				// Vérifier que les données sont présentes (peuvent être nulles pour un tableau vide)
				if response.Data != nil {
					// Vérifier que les données sont un tableau d'événements
					eventsData, ok := response.Data.([]interface{})
					require.True(t, ok, "Les données devraient être un tableau d'événements")

					// Pour les cas avec événements, vérifier qu'il y en a
					if strings.Contains(testCase.CaseName, "avec événements") {
						require.Greater(t, len(eventsData), 0, "Il devrait y avoir au moins un événement")
					}
				} else {
					// Pour les cas sans événements, les données peuvent être nulles
					if strings.Contains(testCase.CaseName, "sans événements") {
						// C'est normal que les données soient nulles pour un tableau vide
					}
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetEventsByWeekRoute teste la route GET de récupération d'événements par semaine avec plusieurs cas
func TestGetEventsByWeekRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie d'événements par semaine avec événements",
			CaseUrl:  "/calendar-event/1/week/2024/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Créer des événements pour la semaine 1 de 2024 (1er janvier 2024)
				result, err := common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement 1 Semaine 1", "Description événement 1", time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), 60, false)
				require.NoError(t, err)
				eventID1, err := result.LastInsertId()
				require.NoError(t, err)

				result, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement 2 Semaine 1", "Description événement 2", time.Date(2024, 1, 5, 14, 30, 0, 0, time.UTC), 90, false)
				require.NoError(t, err)
				eventID2, err := result.LastInsertId()
				require.NoError(t, err)

				// Associer les événements au calendrier
				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID1)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID2)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'événements par semaine sans événements",
			CaseUrl:  "/calendar-event/1/week/2024/2", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'événements par semaine par un utilisateur partagé",
			CaseUrl:  "/calendar-event/1/week/2024/3", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un utilisateur qui aura accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Partager le calendrier avec l'utilisateur
				_, err = common.DB.Exec(`
					INSERT INTO user_calendar (user_id, calendar_id, created_at) 
					VALUES (?, ?, NOW())
				`, user.User.UserID, 1) // Le premier calendrier créé aura l'ID 1
				require.NoError(t, err)

				// Créer un événement pour la semaine 3 de 2024
				result, err := common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement partagé Semaine 3", "Description événement partagé", time.Date(2024, 1, 16, 16, 0, 0, 0, time.UTC), 120, false)
				require.NoError(t, err)
				eventID, err := result.LastInsertId()
				require.NoError(t, err)

				// Associer l'événement au calendrier
				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/calendar-event/1/week/2024/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/calendar-event/1/week/2024/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec token invalide",
			CaseUrl:  "/calendar-event/1/week/2024/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session expirée",
			CaseUrl:  "/calendar-event/1/week/2024/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/calendar-event/1/week/2024/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec calendar_id inexistant",
			CaseUrl:  "/calendar-event/99999/week/2024/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/invalid/week/2024/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de récupération avec année invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/week/invalid/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidYear,
		},
		{
			CaseName: "Échec de récupération avec semaine invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/week/2024/invalid", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidWeekNumber,
		},
		{
			CaseName: "Échec de récupération avec semaine invalide (hors limites)",
			CaseUrl:  "/calendar-event/1/week/2024/54", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidWeekNumber,
		},
		{
			CaseName: "Échec de récupération avec semaine invalide (zéro)",
			CaseUrl:  "/calendar-event/1/week/2024/0", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidWeekNumber,
		},
		{
			CaseName: "Échec de récupération sans accès au calendrier",
			CaseUrl:  "/calendar-event/1/week/2024/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un autre utilisateur sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de récupération d'événements d'un calendrier supprimé",
			CaseUrl:  "/calendar-event/1/week/2024/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer le calendrier (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// Remplacer les IDs dans l'URL si nécessaire
			url := testCase.CaseUrl
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				// Récupérer l'ID du calendrier de l'utilisateur
				var calendarID int
				err := common.DB.QueryRow(`
					SELECT c.calendar_id 
					FROM calendar c
					INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
					WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL
					ORDER BY c.created_at DESC
					LIMIT 1
				`, user.User.UserID).Scan(&calendarID)
				if err == nil {
					// Préserver les paramètres year et week de l'URL originale pour les cas de validation
					if strings.Contains(testCase.CaseName, "invalide") {
						// Pour les cas de validation, garder les paramètres invalides
						urlParts := strings.Split(testCase.CaseUrl, "/")
						if len(urlParts) >= 6 {
							url = "/calendar-event/" + strconv.Itoa(calendarID) + "/week/" + urlParts[4] + "/" + urlParts[5]
						}
					} else {
						// Pour les autres cas, utiliser les paramètres par défaut
						url = "/calendar-event/" + strconv.Itoa(calendarID) + "/week/2024/1"
					}
				}
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header d'authentification si disponible
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+user.SessionToken)
			} else if sessionToken, ok := setupData["sessionToken"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+sessionToken)
			} else if authHeader, ok := setupData["authHeader"].(string); ok {
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}
			}

			// On traite les cas de test un par un.
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
			}

			// Vérifier le message d'erreur
			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError, "Message d'erreur incorrect")
			}

			// Vérifications spécifiques pour les cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")

				// Vérifier que les données sont présentes (peuvent être nulles pour un tableau vide)
				if response.Data != nil {
					// Vérifier que les données sont un tableau d'événements
					eventsData, ok := response.Data.([]interface{})
					require.True(t, ok, "Les données devraient être un tableau d'événements")

					// Pour les cas avec événements, vérifier qu'il y en a
					if strings.Contains(testCase.CaseName, "avec événements") {
						require.Greater(t, len(eventsData), 0, "Il devrait y avoir au moins un événement")
					}
				} else {
					// Pour les cas sans événements, les données peuvent être nulles
					if strings.Contains(testCase.CaseName, "sans événements") {
						// C'est normal que les données soient nulles pour un tableau vide
					}
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetEventsByDayRoute teste la route GET de récupération d'événements par jour avec plusieurs cas
func TestGetEventsByDayRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie d'événements par jour avec événements",
			CaseUrl:  "/calendar-event/1/day/2024/1/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Créer des événements pour le 15 janvier 2024
				result, err := common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement 1 Jour", "Description événement 1", time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), 60, false)
				require.NoError(t, err)
				eventID1, err := result.LastInsertId()
				require.NoError(t, err)

				result, err = common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement 2 Jour", "Description événement 2", time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), 90, false)
				require.NoError(t, err)
				eventID2, err := result.LastInsertId()
				require.NoError(t, err)

				// Associer les événements au calendrier
				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID1)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID2)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'événements par jour sans événements",
			CaseUrl:  "/calendar-event/1/day/2024/1/16", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'événements par jour par un utilisateur partagé",
			CaseUrl:  "/calendar-event/1/day/2024/1/17", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un utilisateur qui aura accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Partager le calendrier avec l'utilisateur
				_, err = common.DB.Exec(`
					INSERT INTO user_calendar (user_id, calendar_id, created_at) 
					VALUES (?, ?, NOW())
				`, user.User.UserID, 1) // Le premier calendrier créé aura l'ID 1
				require.NoError(t, err)

				// Créer un événement pour le 17 janvier 2024
				result, err := common.DB.Exec(`
					INSERT INTO event (title, description, start, duration, canceled, created_at) 
					VALUES (?, ?, ?, ?, ?, NOW())
				`, "Événement partagé Jour", "Description événement partagé", time.Date(2024, 1, 17, 16, 0, 0, 0, time.UTC), 120, false)
				require.NoError(t, err)
				eventID, err := result.LastInsertId()
				require.NoError(t, err)

				// Associer l'événement au calendrier
				_, err = common.DB.Exec(`
					INSERT INTO calendar_event (calendar_id, event_id, created_at) 
					VALUES (?, ?, NOW())
				`, 1, eventID)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/calendar-event/1/day/2024/1/15",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/calendar-event/1/day/2024/1/15",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec token invalide",
			CaseUrl:  "/calendar-event/1/day/2024/1/15",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)

				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session expirée",
			CaseUrl:  "/calendar-event/1/day/2024/1/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/calendar-event/1/day/2024/1/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec calendar_id inexistant",
			CaseUrl:  "/calendar-event/99999/day/2024/1/15",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de récupération avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar-event/invalid/day/2024/1/15",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de récupération avec année invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/day/invalid/1/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidYear,
		},
		{
			CaseName: "Échec de récupération avec mois invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/day/2024/invalid/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidMonth,
		},
		{
			CaseName: "Échec de récupération avec jour invalide (non numérique)",
			CaseUrl:  "/calendar-event/1/day/2024/1/invalid", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidDay,
		},
		{
			CaseName: "Échec de récupération avec mois invalide (hors limites)",
			CaseUrl:  "/calendar-event/1/day/2024/13/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidMonth,
		},
		{
			CaseName: "Échec de récupération avec mois invalide (zéro)",
			CaseUrl:  "/calendar-event/1/day/2024/0/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidMonth,
		},
		{
			CaseName: "Échec de récupération avec jour invalide (hors limites)",
			CaseUrl:  "/calendar-event/1/day/2024/1/32", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidDay,
		},
		{
			CaseName: "Échec de récupération avec jour invalide (zéro)",
			CaseUrl:  "/calendar-event/1/day/2024/1/0", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidDay,
		},
		{
			CaseName: "Échec de récupération sans accès au calendrier",
			CaseUrl:  "/calendar-event/1/day/2024/1/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, true)
				require.NoError(t, err)

				// Créer un autre utilisateur sans accès au calendrier
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				return map[string]interface{}{
					"user":  user,
					"owner": owner,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de récupération d'événements d'un calendrier supprimé",
			CaseUrl:  "/calendar-event/1/day/2024/1/15", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Supprimer le calendrier (soft delete)
				_, err = common.DB.Exec(`
					UPDATE calendar 
					SET deleted_at = NOW() 
					WHERE calendar_id = 1
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// Remplacer les IDs dans l'URL si nécessaire
			url := testCase.CaseUrl
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				// Récupérer l'ID du calendrier de l'utilisateur
				var calendarID int
				err := common.DB.QueryRow(`
					SELECT c.calendar_id 
					FROM calendar c
					INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
					WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL
					ORDER BY c.created_at DESC
					LIMIT 1
				`, user.User.UserID).Scan(&calendarID)
				if err == nil {
					// Préserver les paramètres year, month et day de l'URL originale pour les cas de validation
					if strings.Contains(testCase.CaseName, "invalide") {
						// Pour les cas de validation, garder les paramètres invalides
						urlParts := strings.Split(testCase.CaseUrl, "/")
						if len(urlParts) >= 7 {
							url = "/calendar-event/" + strconv.Itoa(calendarID) + "/day/" + urlParts[4] + "/" + urlParts[5] + "/" + urlParts[6]
						}
					} else {
						// Pour les autres cas, utiliser les paramètres par défaut
						url = "/calendar-event/" + strconv.Itoa(calendarID) + "/day/2024/1/15"
					}
				}
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header d'authentification si disponible
			if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+user.SessionToken)
			} else if sessionToken, ok := setupData["sessionToken"].(string); ok {
				req.Header.Set("Authorization", "Bearer "+sessionToken)
			} else if authHeader, ok := setupData["authHeader"].(string); ok {
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}
			}

			// On traite les cas de test un par un.
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
			}

			// Vérifier le message d'erreur
			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError, "Message d'erreur incorrect")
			}

			// Vérifications spécifiques pour les cas de succès
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")

				// Vérifier que les données sont présentes (peuvent être nulles pour un tableau vide)
				if response.Data != nil {
					// Vérifier que les données sont un tableau d'événements
					eventsData, ok := response.Data.([]interface{})
					require.True(t, ok, "Les données devraient être un tableau d'événements")

					// Pour les cas avec événements, vérifier qu'il y en a
					if strings.Contains(testCase.CaseName, "avec événements") {
						require.Greater(t, len(eventsData), 0, "Il devrait y avoir au moins un événement")
					}
				} else {
					// Pour les cas sans événements, les données peuvent être nulles
					if strings.Contains(testCase.CaseName, "sans événements") {
						// C'est normal que les données soient nulles pour un tableau vide
					}
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}
