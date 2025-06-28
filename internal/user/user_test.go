package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-averroes/internal/common"
	"go-averroes/testutils"

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

// TestRouteExample teste la route POST/PUT d'exemple avec plusieurs cas
func TestRoutePostPutExample(t *testing.T) {
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
			CaseName: "Case name",
			CaseUrl:  "Url",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
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

// TestAddUserRoute teste la route POST de création d'utilisateur avec plusieurs cas
func TestAddUserRoute(t *testing.T) {
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
			CaseName: "Création réussie d'un utilisateur avec données valides",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la création d'un nouvel utilisateur
				return map[string]interface{}{
					"lastname":  "Dupont",
					"firstname": "Jean",
					"email":     "jean.dupont@example.com",
					"password":  "motdepasse123",
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de création avec lastname manquant",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des données
				return map[string]interface{}{
					"firstname": "Jean",
					"email":     "jean.dupont@example.com",
					"password":  "motdepasse123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec firstname manquant",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des données
				return map[string]interface{}{
					"lastname": "Dupont",
					"email":    "jean.dupont@example.com",
					"password": "motdepasse123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec email manquant",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des données
				return map[string]interface{}{
					"lastname":  "Dupont",
					"firstname": "Jean",
					"password":  "motdepasse123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec password manquant",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des données
				return map[string]interface{}{
					"lastname":  "Dupont",
					"firstname": "Jean",
					"email":     "jean.dupont@example.com",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec email invalide",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des données
				return map[string]interface{}{
					"lastname":  "Dupont",
					"firstname": "Jean",
					"email":     "email-invalide",
					"password":  "motdepasse123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec password trop court",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des données
				return map[string]interface{}{
					"lastname":  "Dupont",
					"firstname": "Jean",
					"email":     "jean.dupont@example.com",
					"password":  "123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec email déjà existant",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur existant en base pour tester le conflit
				existingUser, err := testutils.GenerateAuthenticatedUser(false, true)
				require.NoError(t, err)

				// Retourner les données de requête avec l'email existant
				return map[string]interface{}{
					"lastname":     "Dupont",
					"firstname":    "Jean",
					"email":        existingUser.User.Email, // Email déjà existant
					"password":     "motdepasse123",
					"existingUser": existingUser, // Pour le nettoyage après le test
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserAlreadyExists,
		},
		{
			CaseName: "Échec de création avec données JSON invalides",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste le parsing JSON
				return map[string]interface{}{
					"invalid_json": "data",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Création réussie avec caractères spéciaux dans les noms",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des caractères spéciaux
				return map[string]interface{}{
					"lastname":  "O'Connor",
					"firstname": "Jean-Pierre",
					"email":     "jean-pierre.oconnor@example.com",
					"password":  "motdepasse123",
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Création réussie avec email avec tirets et points",
			CaseUrl:  "/user",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation d'email complexe
				return map[string]interface{}{
					"lastname":  "Martin",
					"firstname": "Marie",
					"email":     "marie.martin-test@subdomain.example.co.uk",
					"password":  "motdepasse123",
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateUser,
			ExpectedError:    "",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données de requête
			requestData := testCase.RequestData()

			// Extraire l'utilisateur existant si présent pour le nettoyage
			if userData, ok := requestData["existingUser"]; ok {
				_ = userData.(*testutils.AuthenticatedUser) // Pour le nettoyage après le test
				delete(requestData, "existingUser")         // Supprimer de requestData
			}

			// Convertir les données en JSON
			jsonData, err := json.Marshal(requestData)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", testServer.URL+testCase.CaseUrl, bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header Content-Type
			req.Header.Set("Content-Type", "application/json")

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			// Lire et parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier la réponse selon le cas de test
			if testCase.ExpectedHttpCode == http.StatusCreated {
				require.True(t, response.Success, "La réponse devrait être un succès")
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier que l'utilisateur a bien été créé en base
				if response.Data != nil {
					dataMap, ok := response.Data.(map[string]interface{})
					require.True(t, ok, "La réponse devrait contenir un user_id")

					userID, exists := dataMap["user_id"]
					require.True(t, exists, "La réponse devrait contenir un user_id")
					require.NotNil(t, userID, "Le user_id ne devrait pas être null")
				}
			} else {
				require.False(t, response.Success, "La réponse devrait être un échec")
				require.Equal(t, testCase.ExpectedError, response.Error, "Message d'erreur incorrect")
				require.Empty(t, response.Message, "Pas de message de succès attendu")
			}

			// Nettoyer les données de test
			testutils.PurgeAllTestUsers()
		})
	}
}
