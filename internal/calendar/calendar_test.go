package calendar_test

import (
	"bytes"
	"encoding/json"
	"go-averroes/internal/common"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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

// TestCreateCalendarRoute teste la route POST de création d'un calendrier avec plusieurs cas
func TestCreateCalendarRoute(t *testing.T) {
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
			CaseName: "Création réussie d'un calendrier avec titre et description",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Mon Calendrier Personnel",
						"description": "Calendrier pour mes événements personnels",
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Création réussie d'un calendrier avec titre seulement",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Calendrier Simple",
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de création sans header Authorization",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"title":       "Calendrier Test",
						"description": "Description test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de création avec header Authorization vide",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"title":       "Calendrier Test",
						"description": "Description test",
					},
					"authHeader": "",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de création avec token invalide",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"title":       "Calendrier Test",
						"description": "Description test",
					},
					"authHeader": "Bearer invalid_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec session expirée",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":         user,
					"sessionToken": expiredSessionToken,
					"requestBody": map[string]interface{}{
						"title":       "Calendrier Test",
						"description": "Description test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec session désactivée",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
						"title":       "Calendrier Test",
						"description": "Description test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec titre manquant",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"description": "Description sans titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec titre vide",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "",
						"description": "Description avec titre vide",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec données JSON invalides",
			CaseUrl:  "/calendar",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":        user,
					"requestBody": "invalid_json_data",
				}
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
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.RequestData()

			// Préparer le corps de la requête
			var requestBody []byte
			var err error
			if body, ok := setupData["requestBody"]; ok {
				if bodyStr, isString := body.(string); isString {
					requestBody = []byte(bodyStr)
				} else {
					requestBody, err = json.Marshal(body)
					require.NoError(t, err, "Erreur lors de la sérialisation du corps de la requête")
				}
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", testServer.URL+testCase.CaseUrl, bytes.NewBuffer(requestBody))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers
			req.Header.Set("Content-Type", "application/json")

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
			if testCase.ExpectedHttpCode == http.StatusCreated {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")
				require.NotNil(t, response.Data, "Les données de réponse ne devraient pas être nulles")

				// Vérifier que le calendrier a bien été créé en base
				if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
					// Vérifier que le calendrier existe en base
					var calendarID int
					var title string
					err := common.DB.QueryRow(`
						SELECT c.calendar_id, c.title 
						FROM calendar c
						INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
						WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL
						ORDER BY c.created_at DESC
						LIMIT 1
					`, user.User.UserID).Scan(&calendarID, &title)
					require.NoError(t, err, "Le calendrier devrait être créé en base de données")
					require.Greater(t, calendarID, 0, "L'ID du calendrier devrait être positif")

					// Vérifier que la liaison user_calendar existe
					var userCalendarID int
					err = common.DB.QueryRow(`
						SELECT user_calendar_id 
						FROM user_calendar 
						WHERE user_id = ? AND calendar_id = ? AND deleted_at IS NULL
					`, user.User.UserID, calendarID).Scan(&userCalendarID)
					require.NoError(t, err, "La liaison user_calendar devrait être créée")
					require.Greater(t, userCalendarID, 0, "L'ID de la liaison devrait être positif")
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetCalendarRoute teste la route GET de récupération d'un calendrier par ID avec plusieurs cas
func TestGetCalendarRoute(t *testing.T) {
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
			CaseName: "Récupération réussie d'un calendrier par son propriétaire",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessGetCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'un calendrier partagé par un autre utilisateur",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
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
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessGetCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/calendar/1",
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
			CaseUrl:  "/calendar/1",
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
			CaseUrl:  "/calendar/1",
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
			CaseUrl:  "/calendar/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
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
			CaseUrl:  "/calendar/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
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
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec calendar_id inexistant",
			CaseUrl:  "/calendar/99999",
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
			CaseUrl:  "/calendar/invalid",
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
			CaseName: "Échec de récupération sans accès au calendrier",
			CaseUrl:  "/calendar/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
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
			CaseName: "Échec de récupération d'un calendrier supprimé",
			CaseUrl:  "/calendar/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
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

			// Remplacer l'ID du calendrier dans l'URL si nécessaire
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
					url = "/calendar/" + strconv.Itoa(calendarID)
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

				// Vérifier que les données du calendrier sont présentes
				calendarData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet calendrier")
				require.Contains(t, calendarData, "calendar_id", "Le calendrier devrait avoir un ID")
				require.Contains(t, calendarData, "title", "Le calendrier devrait avoir un titre")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestUpdateCalendarRoute teste la route PUT de modification d'un calendrier avec plusieurs cas
func TestUpdateCalendarRoute(t *testing.T) {
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
			CaseName: "Modification réussie du titre et de la description",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Nouveau Titre du Calendrier",
						"description": "Nouvelle description mise à jour",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie du titre seulement",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title": "Titre Modifié Seulement",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie avec description vide (suppression de la description)",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Titre avec Description Supprimée",
						"description": "",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de modification sans header Authorization",
			CaseUrl:  "/calendar/1",
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
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de modification avec header Authorization vide",
			CaseUrl:  "/calendar/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
					"authHeader": "",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de modification avec token invalide",
			CaseUrl:  "/calendar/1",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un calendrier
				_, err := common.DB.Exec(`
					INSERT INTO calendar (title, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Calendrier Test", "Description test")
				require.NoError(t, err)
				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
					"authHeader": "Bearer invalid_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de modification avec session expirée",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
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
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de modification avec session désactivée",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
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
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de modification avec calendar_id inexistant",
			CaseUrl:  "/calendar/99999",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Échec de modification avec calendar_id invalide (non numérique)",
			CaseUrl:  "/calendar/invalid",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCalendarID,
		},
		{
			CaseName: "Échec de modification sans accès au calendrier",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
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
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Échec de modification avec titre manquant",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"description": "Description sans titre",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de modification avec titre vide",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"title":       "",
						"description": "Description avec titre vide",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de modification avec données JSON invalides",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user":        user,
					"requestBody": "invalid_json_data",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de modification d'un calendrier supprimé",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
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
						"title":       "Nouveau Titre",
						"description": "Nouvelle description",
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
			setupData := testCase.RequestData()

			// Remplacer l'ID du calendrier dans l'URL si nécessaire
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
					url = "/calendar/" + strconv.Itoa(calendarID)
				}
			}

			// Préparer le corps de la requête
			var requestBody []byte
			var err error
			if body, ok := setupData["requestBody"]; ok {
				if bodyStr, isString := body.(string); isString {
					requestBody = []byte(bodyStr)
				} else {
					requestBody, err = json.Marshal(body)
					require.NoError(t, err, "Erreur lors de la sérialisation du corps de la requête")
				}
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("PUT", testServer.URL+url, bytes.NewBuffer(requestBody))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers
			req.Header.Set("Content-Type", "application/json")

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

				// Vérifier que les modifications ont bien été appliquées en base
				if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
					// Récupérer l'ID du calendrier
					var calendarID int
					err := common.DB.QueryRow(`
						SELECT c.calendar_id 
						FROM calendar c
						INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
						WHERE uc.user_id = ? AND c.deleted_at IS NULL AND uc.deleted_at IS NULL
						ORDER BY c.created_at DESC
						LIMIT 1
					`, user.User.UserID).Scan(&calendarID)
					require.NoError(t, err, "Le calendrier devrait exister en base de données")

					// Vérifier que les données ont été mises à jour
					var title, description string
					var updatedAt *time.Time
					err = common.DB.QueryRow(`
						SELECT title, description, updated_at 
						FROM calendar 
						WHERE calendar_id = ?
					`, calendarID).Scan(&title, &description, &updatedAt)
					require.NoError(t, err, "Erreur lors de la vérification des données mises à jour")
					require.NotNil(t, updatedAt, "Le champ updated_at devrait être mis à jour")

					// Vérifier que le titre a été mis à jour
					if requestBody, ok := setupData["requestBody"].(map[string]interface{}); ok {
						if expectedTitle, ok := requestBody["title"].(string); ok {
							require.Equal(t, expectedTitle, title, "Le titre devrait être mis à jour")
						}
					}
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestDeleteCalendarRoute teste la route DELETE de suppression d'un calendrier avec plusieurs cas
func TestDeleteCalendarRoute(t *testing.T) {
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
			CaseName: "Suppression réussie d'un calendrier par son propriétaire",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Suppression réussie d'un calendrier partagé par un utilisateur avec accès",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
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
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteCalendar,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression sans header Authorization",
			CaseUrl:  "/calendar/1",
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
			CaseName: "Échec de suppression avec header Authorization vide",
			CaseUrl:  "/calendar/1",
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
			CaseName: "Échec de suppression avec token invalide",
			CaseUrl:  "/calendar/1",
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
			CaseName: "Échec de suppression avec session expirée",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
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
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
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
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec calendar_id inexistant",
			CaseUrl:  "/calendar/99999",
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
			CaseUrl:  "/calendar/invalid",
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
			CaseName: "Échec de suppression sans accès au calendrier",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur propriétaire du calendrier
				owner, err := testutils.GenerateAuthenticatedUser(false, true, true, false)
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
			CaseName: "Échec de suppression d'un calendrier déjà supprimé",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
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
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrCalendarNotFound,
		},
		{
			CaseName: "Suppression réussie avec événements associés (cascade soft delete)",
			CaseUrl:  "/calendar/1", // Sera remplacé par l'ID réel
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
			ExpectedMessage:  common.MsgSuccessDeleteCalendar,
			ExpectedError:    "",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// Remplacer l'ID du calendrier dans l'URL si nécessaire
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
					url = "/calendar/" + strconv.Itoa(calendarID)
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

				// Vérifier que la suppression a bien été effectuée en base (soft delete)
				if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
					// Récupérer l'ID du calendrier
					var calendarID int
					err := common.DB.QueryRow(`
						SELECT c.calendar_id 
						FROM calendar c
						INNER JOIN user_calendar uc ON c.calendar_id = uc.calendar_id
						WHERE uc.user_id = ? AND uc.deleted_at IS NULL
						ORDER BY c.created_at DESC
						LIMIT 1
					`, user.User.UserID).Scan(&calendarID)

					// Le calendrier devrait être supprimé (soft delete)
					if err == nil {
						// Vérifier que le calendrier est marqué comme supprimé
						var deletedAt *time.Time
						err = common.DB.QueryRow(`
							SELECT deleted_at 
							FROM calendar 
							WHERE calendar_id = ?
						`, calendarID).Scan(&deletedAt)
						require.NoError(t, err, "Le calendrier devrait exister en base")
						require.NotNil(t, deletedAt, "Le calendrier devrait être marqué comme supprimé (deleted_at IS NOT NULL)")

						// Vérifier que la liaison user_calendar est aussi supprimée
						var userCalendarDeletedAt *time.Time
						err = common.DB.QueryRow(`
							SELECT deleted_at 
							FROM user_calendar 
							WHERE user_id = ? AND calendar_id = ?
						`, user.User.UserID, calendarID).Scan(&userCalendarDeletedAt)
						require.NoError(t, err, "La liaison user_calendar devrait exister en base")
						require.NotNil(t, userCalendarDeletedAt, "La liaison user_calendar devrait être marquée comme supprimée")

						// Vérifier que les événements associés sont aussi supprimés (si le test inclut des événements)
						if setupData["user"].(*testutils.AuthenticatedUser).User.UserID == user.User.UserID {
							// Vérifier que les liaisons calendar_event sont supprimées
							var calendarEventDeletedAt *time.Time
							err = common.DB.QueryRow(`
								SELECT deleted_at 
								FROM calendar_event 
								WHERE calendar_id = ?
								LIMIT 1
							`, calendarID).Scan(&calendarEventDeletedAt)
							if err == nil {
								require.NotNil(t, calendarEventDeletedAt, "Les liaisons calendar_event devraient être marquées comme supprimées")
							}
						}
					}
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}
