package calendar_test

import (
	"bytes"
	"encoding/json"
	"go-averroes/internal/common"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"os"
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

// TestRouteExample teste la route GET/DELETE d'exemple avec plusieurs cas
func TestRouteGetDeleteExample(t *testing.T) {
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
			CaseName: "Case name",
			CaseUrl:  "Url",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				return map[string]interface{}{}
			},
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
			testutils.PurgeAllTestUsers()
		})
	}
}
