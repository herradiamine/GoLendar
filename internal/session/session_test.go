package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// TestLoginRoute teste la route POST de connexion avec plusieurs cas
func TestLoginRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		RequestData      func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Connexion réussie avec utilisateur valide",
			CaseUrl:  "/auth/login",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur de test en base
				user, err := testutils.GenerateAuthenticatedUser(false, true)
				require.NoError(t, err)

				// Retourner les données de requête avec les credentials de l'utilisateur créé
				return map[string]interface{}{
					"email":    user.User.Email,
					"password": user.Password,
					"user":     user, // Pour le nettoyage après le test
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessLogin,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de connexion avec email inexistant",
			CaseUrl:  "/auth/login",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste avec un email qui n'existe pas
				return map[string]interface{}{
					"email":    "nonexistent@example.com",
					"password": "UserPassword123!",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCredentials,
		},
		{
			CaseName: "Échec de connexion avec mot de passe incorrect",
			CaseUrl:  "/auth/login",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur de test en base
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				require.NoError(t, err)

				// Retourner les données de requête avec un mot de passe incorrect
				return map[string]interface{}{
					"email":    user.User.Email,
					"password": "WrongPassword123!",
					"user":     user, // Pour le nettoyage après le test
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidCredentials,
		},
		{
			CaseName: "Échec de connexion avec email manquant",
			CaseUrl:  "/auth/login",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des champs requis
				return map[string]interface{}{
					"password": "UserPassword123!",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de connexion avec mot de passe manquant",
			CaseUrl:  "/auth/login",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des champs requis
				return map[string]interface{}{
					"email": "test@example.com",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de connexion avec email invalide",
			CaseUrl:  "/auth/login",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation du format email
				return map[string]interface{}{
					"email":    "invalid-email",
					"password": "UserPassword123!",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de connexion avec données vides",
			CaseUrl:  "/auth/login",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des données vides
				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			requestData := testCase.RequestData()

			// Extraire les données de requête et les données de nettoyage
			email, emailExists := requestData["email"].(string)
			password, passwordExists := requestData["password"].(string)

			// Préparer la requête JSON (seulement email et password)
			jsonRequest := map[string]interface{}{}
			if emailExists {
				jsonRequest["email"] = email
			}
			if passwordExists {
				jsonRequest["password"] = password
			}

			jsonData, err := json.Marshal(jsonRequest)
			require.NoError(t, err)

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", testServer.URL+testCase.CaseUrl, bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Exécuter la requête
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode,
				fmt.Sprintf("Code HTTP attendu: %d, reçu: %d", testCase.ExpectedHttpCode, resp.StatusCode))

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Vérifier les champs de la réponse
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait être un succès")
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier que les données de réponse contiennent les informations attendues
				if response.Data != nil {
					// Convertir les données en map pour vérification
					dataMap, ok := response.Data.(map[string]interface{})
					require.True(t, ok, "Les données de réponse devraient être un objet")

					// Vérifier la présence des champs attendus
					require.Contains(t, dataMap, "session_token", "Token de session manquant")
					require.Contains(t, dataMap, "refresh_token", "Refresh token manquant")
					require.Contains(t, dataMap, "expires_at", "Date d'expiration manquante")
					require.Contains(t, dataMap, "user", "Données utilisateur manquantes")
					require.Contains(t, dataMap, "roles", "Rôles manquants")
				}
			} else {
				require.False(t, response.Success, "La réponse devrait être un échec")
				require.Equal(t, testCase.ExpectedError, response.Error, "Message d'erreur incorrect")
				require.Empty(t, response.Message, "Pas de message de succès attendu")
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
		})
	}
}
