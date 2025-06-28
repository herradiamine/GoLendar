package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// TestGetUserMeRoute teste la route GET de récupération des données de l'utilisateur connecté avec plusieurs cas
func TestGetUserMeRoute(t *testing.T) {
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
			CaseName: "Récupération réussie des données de l'utilisateur connecté",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"user": user, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + user.SessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Aucune préparation nécessaire, on teste l'absence de header
				return map[string]interface{}{
					"_headers": map[string]string{
						"Content-Type": "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Aucune préparation nécessaire, on teste le header vide
				return map[string]interface{}{
					"_headers": map[string]string{
						"Authorization": "",
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec format de token invalide",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Aucune préparation nécessaire, on teste le format invalide
				return map[string]interface{}{
					"_headers": map[string]string{
						"Authorization": "InvalidFormat token123",
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec token inexistant",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Aucune préparation nécessaire, on teste avec un token qui n'existe pas
				return map[string]interface{}{
					"_headers": map[string]string{
						"Authorization": "Bearer invalid_token_12345",
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec token expiré",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true)
				require.NoError(t, err)

				// Créer une session expirée manuellement
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour) // Expirée depuis 1 heure
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"user": user, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + expiredSessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				require.NoError(t, err)

				// Désactiver la session manuellement
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE, updated_at = NOW() 
					WHERE session_token = ?
				`, user.SessionToken)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"user": user, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + user.SessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données de test
			setupData := testCase.SetupData()

			// Extraire les headers si présents
			var headers map[string]string
			if headersData, ok := setupData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(setupData, "_headers") // Supprimer les headers des données de setup
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+testCase.CaseUrl, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			if headers != nil {
				for key, value := range headers {
					req.Header.Set(key, value)
				}
			}

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
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait être un succès")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier que les données utilisateur sont présentes
				require.NotNil(t, response.Data, "Les données utilisateur devraient être présentes")

				// Vérifier la structure des données utilisateur
				userData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet utilisateur")

				// Vérifier les champs obligatoires
				require.Contains(t, userData, "user_id", "Le user_id devrait être présent")
				require.Contains(t, userData, "lastname", "Le lastname devrait être présent")
				require.Contains(t, userData, "firstname", "Le firstname devrait être présent")
				require.Contains(t, userData, "email", "L'email devrait être présent")
				require.Contains(t, userData, "created_at", "Le created_at devrait être présent")

				// Vérifier que les données correspondent à l'utilisateur authentifié
				if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
					require.Equal(t, float64(user.User.UserID), userData["user_id"], "Le user_id devrait correspondre")
					require.Equal(t, user.User.Lastname, userData["lastname"], "Le lastname devrait correspondre")
					require.Equal(t, user.User.Firstname, userData["firstname"], "Le firstname devrait correspondre")
					require.Equal(t, user.User.Email, userData["email"], "L'email devrait correspondre")
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
