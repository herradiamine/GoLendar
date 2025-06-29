package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
				existingUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
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
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
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
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			for key, value := range headers {
				req.Header.Set(key, value)
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

// TestUpdateUserMeRoute teste la route PUT de modification des données de l'utilisateur connecté avec plusieurs cas
func TestUpdateUserMeRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		RequestData      func() map[string]interface{}
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Modification réussie du lastname",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie du firstname",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"firstname": "NouveauPrénom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie de l'email",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"email": "nouveau.email@example.com",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie du mot de passe",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"password": "nouveaumotdepasse123",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie de plusieurs champs",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname":  "NouveauNom",
					"firstname": "NouveauPrénom",
					"email":     "nouveau.email@example.com",
					"password":  "nouveaumotdepasse123",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de modification sans header Authorization",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
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
			CaseName: "Échec de modification avec token invalide",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Aucune préparation nécessaire, on teste le token invalide
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
			CaseName: "Échec de modification avec email invalide",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"email": "email-invalide",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEmailFormat,
		},
		{
			CaseName: "Échec de modification avec password trop court",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"password": "123",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrPasswordTooShort,
		},
		{
			CaseName: "Échec de modification avec email déjà existant",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"email": "email.existant@example.com", // Sera remplacé par l'email de l'utilisateur existant
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer deux utilisateurs cibles
				user2, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"user":  user,  // Utilisateur qui fait la modification
					"user2": user2, // Utilisateur avec l'email existant
					"_headers": map[string]string{
						"Authorization": "Bearer " + user.SessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserAlreadyExists,
		},
		{
			CaseName: "Modification réussie avec données JSON vides (aucune modification)",
			CaseUrl:  "/user/me",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données de test
			setupData := testCase.SetupData()
			requestData := testCase.RequestData()

			// Extraire les headers si présents
			var headers map[string]string
			if headersData, ok := setupData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(setupData, "_headers") // Supprimer les headers des données de setup
			}

			// Gérer le cas spécial de l'email déjà existant
			if testCase.CaseName == "Échec de modification avec email déjà existant" {
				if user2, ok := setupData["user2"].(*testutils.AuthenticatedUser); ok {
					requestData["email"] = user2.User.Email
				}
			}

			// Convertir les données en JSON
			jsonData, err := json.Marshal(requestData)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("PUT", testServer.URL+testCase.CaseUrl, bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			for key, value := range headers {
				req.Header.Set(key, value)
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
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier que les modifications ont bien été appliquées en base
				if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
					// Récupérer les données mises à jour depuis la base
					var updatedUser common.User
					err := common.DB.QueryRow(`
						SELECT user_id, lastname, firstname, email, created_at, updated_at, deleted_at
						FROM user 
						WHERE user_id = ? AND deleted_at IS NULL
					`, user.User.UserID).Scan(
						&updatedUser.UserID,
						&updatedUser.Lastname,
						&updatedUser.Firstname,
						&updatedUser.Email,
						&updatedUser.CreatedAt,
						&updatedUser.UpdatedAt,
						&updatedUser.DeletedAt,
					)
					require.NoError(t, err, "Erreur lors de la récupération de l'utilisateur mis à jour")

					// Vérifier les champs modifiés
					if newLastname, ok := requestData["lastname"].(string); ok {
						require.Equal(t, newLastname, updatedUser.Lastname, "Le lastname devrait être mis à jour")
					}
					if newFirstname, ok := requestData["firstname"].(string); ok {
						require.Equal(t, newFirstname, updatedUser.Firstname, "Le firstname devrait être mis à jour")
					}
					if newEmail, ok := requestData["email"].(string); ok {
						require.Equal(t, newEmail, updatedUser.Email, "L'email devrait être mis à jour")
					}

					// Vérifier que updated_at a été mis à jour
					require.NotNil(t, updatedUser.UpdatedAt, "Le champ updated_at devrait être mis à jour")
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

// TestDeleteUserMeRoute teste la route DELETE de suppression du compte utilisateur connecté avec plusieurs cas
func TestDeleteUserMeRoute(t *testing.T) {
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
			CaseName: "Suppression réussie du compte utilisateur",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			ExpectedMessage:  common.MsgSuccessUserDelete,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression sans header Authorization",
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
			CaseName: "Échec de suppression avec header Authorization vide",
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
			CaseName: "Échec de suppression avec format de token invalide",
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
			CaseName: "Échec de suppression avec token inexistant",
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
			CaseName: "Échec de suppression avec token expiré",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
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
			CaseName: "Échec de suppression avec session désactivée",
			CaseUrl:  "/user/me",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			req, err := http.NewRequest("DELETE", testServer.URL+testCase.CaseUrl, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			for key, value := range headers {
				req.Header.Set(key, value)
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
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier que l'utilisateur a bien été supprimé (soft delete) en base
				if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
					// Vérifier que l'utilisateur est maintenant supprimé (deleted_at IS NOT NULL)
					var deletedAt *time.Time
					err := common.DB.QueryRow(`
						SELECT deleted_at 
						FROM user 
						WHERE user_id = ?
					`, user.User.UserID).Scan(&deletedAt)
					require.NoError(t, err, "Erreur lors de la vérification de la suppression de l'utilisateur")
					require.NotNil(t, deletedAt, "L'utilisateur devrait être supprimé (deleted_at IS NOT NULL)")

					// Vérifier que le mot de passe est également supprimé
					err = common.DB.QueryRow(`
						SELECT deleted_at 
						FROM user_password 
						WHERE user_id = ?
					`, user.User.UserID).Scan(&deletedAt)
					require.NoError(t, err, "Erreur lors de la vérification de la suppression du mot de passe")
					require.NotNil(t, deletedAt, "Le mot de passe devrait être supprimé (deleted_at IS NOT NULL)")

					// Note: Les sessions ne sont pas automatiquement supprimées lors de la suppression d'un utilisateur
					// car elles peuvent être gérées séparément (déconnexion, expiration, etc.)
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

// TestGetUserByIDRoute teste la route GET de récupération d'un utilisateur par ID (admin) avec plusieurs cas
func TestGetUserByIDRoute(t *testing.T) {
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
			CaseName: "Récupération réussie d'un utilisateur par ID par un admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à récupérer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur cible à récupérer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Content-Type": "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur cible à récupérer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "",
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec token invalide",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur cible à récupérer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer invalid_token_12345",
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération sans rôle admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur normal (non-admin) authentifié avec session active en base
				normalUser, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à récupérer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"normalUser": normalUser, // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + normalUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de récupération avec user_id inexistant",
			CaseUrl:  "/user/99999",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"adminUser": adminUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "99999", // ID inexistant
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de récupération avec user_id invalide (non numérique)",
			CaseUrl:  "/user/invalid",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"adminUser": adminUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "invalid", // ID invalide
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidUserID,
		},
		{
			CaseName: "Échec de récupération avec token expiré",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin avec session expirée en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(false, true, false, false)
				require.NoError(t, err)

				// Créer une session expirée manuellement
				expiredSessionToken, _, _, err := testutils.CreateUserSession(adminUser.User.UserID, -1*time.Hour) // Expirée depuis 1 heure
				require.NoError(t, err)

				// Créer un utilisateur cible à récupérer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + expiredSessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Désactiver la session manuellement
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE, updated_at = NOW() 
					WHERE session_token = ?
				`, adminUser.SessionToken)
				require.NoError(t, err)

				// Créer un utilisateur cible à récupérer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
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

			// Extraire les headers et paramètres URL si présents
			var headers map[string]string
			var urlParams map[string]string
			if headersData, ok := setupData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(setupData, "_headers") // Supprimer les headers des données de setup
			}
			if urlParamsData, ok := setupData["_urlParams"]; ok {
				urlParams = urlParamsData.(map[string]string)
				delete(setupData, "_urlParams") // Supprimer les paramètres URL des données de setup
			}

			// Construire l'URL avec les paramètres
			url := testCase.CaseUrl
			if userID, ok := urlParams["user_id"]; ok {
				url = "/user/" + userID
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			for key, value := range headers {
				req.Header.Set(key, value)
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

				// Vérifier que les données correspondent à l'utilisateur admin authentifié
				if adminUser, ok := setupData["adminUser"].(*testutils.AuthenticatedUser); ok {
					require.Equal(t, float64(adminUser.User.UserID), userData["user_id"], "Le user_id devrait correspondre")
					require.Equal(t, adminUser.User.Lastname, userData["lastname"], "Le lastname devrait correspondre")
					require.Equal(t, adminUser.User.Firstname, userData["firstname"], "Le firstname devrait correspondre")
					require.Equal(t, adminUser.User.Email, userData["email"], "L'email devrait correspondre")
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

// TestUpdateUserByIDRoute teste la route PUT de modification d'un utilisateur par ID (admin) avec plusieurs cas
func TestUpdateUserByIDRoute(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		RequestData      func() map[string]interface{}
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Modification réussie du lastname d'un utilisateur par un admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie du firstname d'un utilisateur par un admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"firstname": "NouveauPrénom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie de l'email d'un utilisateur par un admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"email": "nouveau.email@example.com",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie du mot de passe d'un utilisateur par un admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"password": "nouveaumotdepasse123",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Modification réussie de plusieurs champs d'un utilisateur par un admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname":  "NouveauNom",
					"firstname": "NouveauPrénom",
					"password":  "nouveaumotdepasse123",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de modification sans header Authorization",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
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
			CaseName: "Échec de modification avec token invalide",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer invalid_token_12345",
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de modification sans rôle admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur normal (non-admin) authentifié avec session active en base
				normalUser, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"normalUser": normalUser, // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + normalUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de modification avec user_id inexistant",
			CaseUrl:  "/user/99999",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"adminUser": adminUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "99999", // ID inexistant
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de modification avec user_id invalide (non numérique)",
			CaseUrl:  "/user/invalid",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"lastname": "NouveauNom",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"adminUser": adminUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "invalid", // ID invalide
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidUserID,
		},
		{
			CaseName: "Échec de modification avec email invalide",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"email": "email-invalide",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEmailFormat,
		},
		{
			CaseName: "Échec de modification avec password trop court",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"password": "123",
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrPasswordTooShort,
		},
		{
			CaseName: "Échec de modification avec email déjà existant",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"email": "email.existant@example.com", // Sera remplacé par l'email de l'utilisateur existant
				}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer deux utilisateurs cibles
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				existingUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":    adminUser,    // Pour le nettoyage après le test
					"targetUser":   targetUser,   // Utilisateur à modifier
					"existingUser": existingUser, // Utilisateur avec l'email existant
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserAlreadyExists,
		},
		{
			CaseName: "Modification réussie avec données JSON vides (aucune modification)",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{}
			},
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à modifier
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUpdateUser,
			ExpectedError:    "",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Préparer les données de test
			setupData := testCase.SetupData()
			requestData := testCase.RequestData()

			// Extraire les headers et paramètres URL si présents
			var headers map[string]string
			var urlParams map[string]string
			if headersData, ok := setupData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(setupData, "_headers") // Supprimer les headers des données de setup
			}
			if urlParamsData, ok := setupData["_urlParams"]; ok {
				urlParams = urlParamsData.(map[string]string)
				delete(setupData, "_urlParams") // Supprimer les paramètres URL des données de setup
			}

			// Gérer le cas spécial de l'email déjà existant
			if testCase.CaseName == "Échec de modification avec email déjà existant" {
				if existingUser, ok := setupData["existingUser"].(*testutils.AuthenticatedUser); ok {
					requestData["email"] = existingUser.User.Email
				}
			}

			// Construire l'URL avec les paramètres
			url := testCase.CaseUrl
			if userID, ok := urlParams["user_id"]; ok {
				url = "/user/" + userID
			}

			// Convertir les données en JSON
			jsonData, err := json.Marshal(requestData)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("PUT", testServer.URL+url, bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			for key, value := range headers {
				req.Header.Set(key, value)
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
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier que les modifications ont bien été appliquées en base
				if adminUser, ok := setupData["adminUser"].(*testutils.AuthenticatedUser); ok {
					// Récupérer les données mises à jour depuis la base
					var updatedUser common.User
					err := common.DB.QueryRow(`
						SELECT user_id, lastname, firstname, email, created_at, updated_at, deleted_at
						FROM user 
						WHERE user_id = ? AND deleted_at IS NULL
					`, adminUser.User.UserID).Scan(
						&updatedUser.UserID,
						&updatedUser.Lastname,
						&updatedUser.Firstname,
						&updatedUser.Email,
						&updatedUser.CreatedAt,
						&updatedUser.UpdatedAt,
						&updatedUser.DeletedAt,
					)
					require.NoError(t, err, "Erreur lors de la récupération de l'utilisateur mis à jour")

					// Vérifier les champs modifiés
					if newLastname, ok := requestData["lastname"].(string); ok {
						require.Equal(t, newLastname, updatedUser.Lastname, "Le lastname devrait être mis à jour")
					}
					if newFirstname, ok := requestData["firstname"].(string); ok {
						require.Equal(t, newFirstname, updatedUser.Firstname, "Le firstname devrait être mis à jour")
					}
					if newEmail, ok := requestData["email"].(string); ok {
						require.Equal(t, newEmail, updatedUser.Email, "L'email devrait être mis à jour")
					}

					// Vérifier que updated_at a été mis à jour (seulement si des modifications ont été faites)
					if len(requestData) > 0 {
						require.NotNil(t, updatedUser.UpdatedAt, "Le champ updated_at devrait être mis à jour")
					}
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

// TestDeleteUserByIDRoute teste la route DELETE de suppression d'un utilisateur par ID (admin) avec plusieurs cas
func TestDeleteUserByIDRoute(t *testing.T) {
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
			CaseName: "Suppression réussie d'un utilisateur par ID par un admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à supprimer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserDelete,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression sans header Authorization",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur cible à supprimer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Content-Type": "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de suppression avec header Authorization vide",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur cible à supprimer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "",
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de suppression avec token invalide",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur cible à supprimer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer invalid_token_12345",
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression sans rôle admin",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur normal (non-admin) authentifié avec session active en base
				normalUser, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer un utilisateur cible à supprimer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"normalUser": normalUser, // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + normalUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de suppression avec user_id inexistant",
			CaseUrl:  "/user/99999",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"adminUser": adminUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "99999", // ID inexistant
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec de suppression avec user_id invalide (non numérique)",
			CaseUrl:  "/user/invalid",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"adminUser": adminUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "invalid", // ID invalide
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidUserID,
		},
		{
			CaseName: "Échec de suppression avec token expiré",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin avec session expirée en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(false, true, false, false)
				require.NoError(t, err)

				// Créer une session expirée manuellement
				expiredSessionToken, _, _, err := testutils.CreateUserSession(adminUser.User.UserID, -1*time.Hour) // Expirée depuis 1 heure
				require.NoError(t, err)

				// Créer un utilisateur cible à supprimer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + expiredSessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de suppression avec session désactivée",
			CaseUrl:  "/user/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur admin avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Désactiver la session manuellement
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE, updated_at = NOW() 
					WHERE session_token = ?
				`, adminUser.SessionToken)
				require.NoError(t, err)

				// Créer un utilisateur cible à supprimer
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs et les headers
				return map[string]interface{}{
					"adminUser":  adminUser,  // Pour le nettoyage après le test
					"targetUser": targetUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
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

			// Extraire les headers et paramètres URL si présents
			var headers map[string]string
			var urlParams map[string]string
			if headersData, ok := setupData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(setupData, "_headers") // Supprimer les headers des données de setup
			}
			if urlParamsData, ok := setupData["_urlParams"]; ok {
				urlParams = urlParamsData.(map[string]string)
				delete(setupData, "_urlParams") // Supprimer les paramètres URL des données de setup
			}

			// Construire l'URL avec les paramètres
			url := testCase.CaseUrl
			if userID, ok := urlParams["user_id"]; ok {
				url = "/user/" + userID
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("DELETE", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			for key, value := range headers {
				req.Header.Set(key, value)
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
				require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier que l'utilisateur a bien été supprimé (soft delete) en base
				if adminUser, ok := setupData["adminUser"].(*testutils.AuthenticatedUser); ok {
					// Vérifier que l'utilisateur admin est maintenant supprimé (deleted_at IS NOT NULL)
					var deletedAt *time.Time
					err := common.DB.QueryRow(`
						SELECT deleted_at 
						FROM user 
						WHERE user_id = ?
					`, adminUser.User.UserID).Scan(&deletedAt)
					require.NoError(t, err, "Erreur lors de la vérification de la suppression de l'utilisateur")
					require.NotNil(t, deletedAt, "L'utilisateur devrait être supprimé (deleted_at IS NOT NULL)")

					// Vérifier que le mot de passe est également supprimé
					err = common.DB.QueryRow(`
						SELECT deleted_at 
						FROM user_password 
						WHERE user_id = ?
					`, adminUser.User.UserID).Scan(&deletedAt)
					require.NoError(t, err, "Erreur lors de la vérification de la suppression du mot de passe")
					require.NotNil(t, deletedAt, "Le mot de passe devrait être supprimé (deleted_at IS NOT NULL)")

					// Note: Les sessions ne sont pas automatiquement supprimées lors de la suppression d'un utilisateur
					// car elles peuvent être gérées séparément (déconnexion, expiration, etc.)
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

// TestGetUserByIDWithRolesRoute teste la route GET de récupération d'un utilisateur par ID avec ses rôles (admin) avec plusieurs cas
func TestGetUserByIDWithRolesRoute(t *testing.T) {
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		SetupData        func() map[string]interface{}
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Récupération réussie d'un utilisateur avec ses rôles par un admin",
			CaseUrl:  "/user/1/with-roles", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"adminUser":  adminUser,
					"targetUser": targetUser,
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec sans header Authorization",
			CaseUrl:  "/user/1/with-roles",
			SetupData: func() map[string]interface{} {
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"targetUser": targetUser,
					"_headers": map[string]string{
						"Content-Type": "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec avec header Authorization vide",
			CaseUrl:  "/user/1/with-roles",
			SetupData: func() map[string]interface{} {
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"targetUser": targetUser,
					"_headers": map[string]string{
						"Authorization": "",
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec avec token invalide",
			CaseUrl:  "/user/1/with-roles",
			SetupData: func() map[string]interface{} {
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"targetUser": targetUser,
					"_headers": map[string]string{
						"Authorization": "Bearer invalid_token_12345",
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec sans rôle admin",
			CaseUrl:  "/user/1/with-roles",
			SetupData: func() map[string]interface{} {
				normalUser, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"normalUser": normalUser,
					"targetUser": targetUser,
					"_headers": map[string]string{
						"Authorization": "Bearer " + normalUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec avec user_id inexistant",
			CaseUrl:  "/user/99999/with-roles",
			SetupData: func() map[string]interface{} {
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"adminUser": adminUser,
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "99999",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotFound,
		},
		{
			CaseName: "Échec avec user_id invalide (non numérique)",
			CaseUrl:  "/user/invalid/with-roles",
			SetupData: func() map[string]interface{} {
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"adminUser": adminUser,
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": "invalid",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidUserID,
		},
		{
			CaseName: "Échec avec token expiré",
			CaseUrl:  "/user/1/with-roles",
			SetupData: func() map[string]interface{} {
				adminUser, err := testutils.GenerateAuthenticatedAdmin(false, true, false, false)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(adminUser.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"adminUser":  adminUser,
					"targetUser": targetUser,
					"_headers": map[string]string{
						"Authorization": "Bearer " + expiredSessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec avec session désactivée",
			CaseUrl:  "/user/1/with-roles",
			SetupData: func() map[string]interface{} {
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE, updated_at = NOW() 
					WHERE session_token = ?
				`, adminUser.SessionToken)
				require.NoError(t, err)
				targetUser, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)
				return map[string]interface{}{
					"adminUser":  adminUser,
					"targetUser": targetUser,
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
					"_urlParams": map[string]string{
						"user_id": strconv.Itoa(targetUser.User.UserID),
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			setupData := testCase.SetupData()
			var headers map[string]string
			var urlParams map[string]string
			if headersData, ok := setupData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(setupData, "_headers")
			}
			if urlParamsData, ok := setupData["_urlParams"]; ok {
				urlParams = urlParamsData.(map[string]string)
				delete(setupData, "_urlParams")
			}
			url := testCase.CaseUrl
			if userID, ok := urlParams["user_id"]; ok {
				url = "/user/" + userID + "/with-roles"
			}
			req, err := http.NewRequest("GET", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")
			for key, value := range headers {
				req.Header.Set(key, value)
			}
			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()
			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")
			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait être un succès")
				require.Empty(t, response.Error, "Pas d'erreur attendue")
				// Vérifier la structure des données utilisateur + rôles
				userData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet utilisateur")
				require.Contains(t, userData, "user_id")
				require.Contains(t, userData, "lastname")
				require.Contains(t, userData, "firstname")
				require.Contains(t, userData, "email")
				require.Contains(t, userData, "created_at")
				require.Contains(t, userData, "roles")
				// Vérifier que les rôles sont bien un tableau
				_, ok = userData["roles"].([]interface{})
				require.True(t, ok, "Les rôles doivent être un tableau")
			} else {
				require.False(t, response.Success, "La réponse devrait être un échec")
				require.Equal(t, testCase.ExpectedError, response.Error, "Message d'erreur incorrect")
				require.Empty(t, response.Message, "Pas de message de succès attendu")
			}
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetAuthMeRoute teste la route GET /auth/me de récupération des informations de l'utilisateur authentifié avec ses rôles
func TestGetAuthMeRoute(t *testing.T) {
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
			CaseName: "Récupération réussie des informations de l'utilisateur authentifié avec ses rôles",
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur authentifié avec session active et rôles en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
			CaseName: "Récupération réussie d'un utilisateur admin avec ses rôles",
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur admin authentifié avec session active en base
				adminUser, err := testutils.GenerateAuthenticatedAdmin(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"adminUser": adminUser, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + adminUser.SessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Récupération réussie d'un utilisateur sans rôles",
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur normal sans rôles assignés
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Supprimer tous les rôles de l'utilisateur
				_, err = common.DB.Exec("DELETE FROM user_roles WHERE user_id = ?", user.User.UserID)
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
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
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
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
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
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
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
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
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
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
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
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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
		{
			CaseName: "Échec de récupération avec utilisateur supprimé",
			CaseUrl:  "/auth/me",
			SetupData: func() map[string]interface{} {
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Supprimer l'utilisateur (soft delete)
				_, err = common.DB.Exec("UPDATE user SET deleted_at = NOW() WHERE user_id = ?", user.User.UserID)
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
			setupData := testCase.SetupData()
			var headers map[string]string
			if headersData, ok := setupData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(setupData, "_headers")
			}

			req, err := http.NewRequest("GET", testServer.URL+testCase.CaseUrl, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")
			for key, value := range headers {
				req.Header.Set(key, value)
			}

			resp, err := testClient.Do(req)
			require.NoError(t, err, "Erreur lors de l'exécution de la requête")
			defer resp.Body.Close()

			var response common.JSONResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			require.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode, "Code de statut HTTP incorrect")

			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait être un succès")
				require.Empty(t, response.Error, "Pas d'erreur attendue")

				// Vérifier la structure des données utilisateur + rôles
				userData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet utilisateur")

				// Vérifier les champs de l'utilisateur
				require.Contains(t, userData, "user_id")
				require.Contains(t, userData, "lastname")
				require.Contains(t, userData, "firstname")
				require.Contains(t, userData, "email")
				require.Contains(t, userData, "created_at")
				require.Contains(t, userData, "roles")

				// Vérifier que les rôles sont bien un tableau
				roles, ok := userData["roles"].([]interface{})
				require.True(t, ok, "Les rôles doivent être un tableau")

				// Si l'utilisateur a des rôles, vérifier leur structure
				if len(roles) > 0 {
					firstRole, ok := roles[0].(map[string]interface{})
					require.True(t, ok, "Le premier rôle doit être un objet")
					require.Contains(t, firstRole, "role_id")
					require.Contains(t, firstRole, "name")
					require.Contains(t, firstRole, "description")
					require.Contains(t, firstRole, "created_at")
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
