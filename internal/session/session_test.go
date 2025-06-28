package session_test

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
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
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
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
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

// TestRefreshTokenRoute teste la route POST de rafraîchissement de token avec plusieurs cas
func TestRefreshTokenRoute(t *testing.T) {
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
			CaseName: "Rafraîchissement réussi avec refresh token valide",
			CaseUrl:  "/auth/refresh",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de requête avec le refresh token de l'utilisateur créé
				return map[string]interface{}{
					"refresh_token": user.RefreshToken,
					"user":          user, // Pour le nettoyage après le test
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessRefreshToken,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de rafraîchissement avec refresh token inexistant",
			CaseUrl:  "/auth/refresh",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste avec un refresh token qui n'existe pas
				return map[string]interface{}{
					"refresh_token": "invalid_refresh_token_12345",
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de rafraîchissement avec refresh token expiré",
			CaseUrl:  "/auth/refresh",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer une session expirée manuellement
				_, expiredRefreshToken, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour) // Expirée depuis 1 heure
				require.NoError(t, err)

				// Retourner les données de requête avec le refresh token expiré
				return map[string]interface{}{
					"refresh_token": expiredRefreshToken,
					"user":          user, // Pour le nettoyage après le test
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionExpired,
		},
		{
			CaseName: "Échec de rafraîchissement avec refresh token manquant",
			CaseUrl:  "/auth/refresh",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des champs requis
				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de rafraîchissement avec refresh token vide",
			CaseUrl:  "/auth/refresh",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Aucune préparation nécessaire, on teste la validation des champs requis
				return map[string]interface{}{
					"refresh_token": "",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de rafraîchissement avec session désactivée",
			CaseUrl:  "/auth/refresh",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Désactiver la session manuellement
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE, updated_at = NOW() 
					WHERE refresh_token = ?
				`, user.RefreshToken)
				require.NoError(t, err)

				// Retourner les données de requête avec le refresh token de la session désactivée
				return map[string]interface{}{
					"refresh_token": user.RefreshToken,
					"user":          user, // Pour le nettoyage après le test
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
			// Préparer les données de requête
			requestData := testCase.RequestData()

			// Créer le body de la requête JSON
			jsonBody, err := json.Marshal(requestData)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", testServer.URL+testCase.CaseUrl, bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
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
					require.Contains(t, dataMap, "session_token", "Nouveau token de session manquant")
					require.Contains(t, dataMap, "expires_at", "Nouvelle date d'expiration manquante")

					// Vérifier que le nouveau session token est différent de l'ancien
					newSessionToken, ok := dataMap["session_token"].(string)
					require.True(t, ok, "Le nouveau session token devrait être une chaîne")
					require.NotEmpty(t, newSessionToken, "Le nouveau session token ne devrait pas être vide")

					// Vérifier que la nouvelle date d'expiration est dans le futur
					expiresAtStr, ok := dataMap["expires_at"].(string)
					require.True(t, ok, "La date d'expiration devrait être une chaîne")
					newExpiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
					require.NoError(t, err, "Erreur lors du parsing de la date d'expiration")
					require.True(t, newExpiresAt.After(time.Now()), "La nouvelle date d'expiration devrait être dans le futur")
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

// TestLogoutRoute teste la route POST de déconnexion avec plusieurs cas
func TestLogoutRoute(t *testing.T) {
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
			CaseName: "Déconnexion réussie avec token de session valide",
			CaseUrl:  "/auth/logout",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Retourner les données de requête avec l'utilisateur et les headers nécessaires
				return map[string]interface{}{
					"user": user, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + user.SessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessLogout,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de déconnexion sans header Authorization",
			CaseUrl:  "/auth/logout",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
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
			CaseName: "Échec de déconnexion avec header Authorization vide",
			CaseUrl:  "/auth/logout",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
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
			ExpectedError:    common.ErrUserNotAuthenticated, // Le middleware retourne cette erreur pour un header vide
		},
		{
			CaseName: "Échec de déconnexion avec format de token invalide",
			CaseUrl:  "/auth/logout",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
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
			CaseName: "Échec de déconnexion avec token inexistant",
			CaseUrl:  "/auth/logout",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
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
			CaseName: "Échec de déconnexion avec token expiré",
			CaseUrl:  "/auth/logout",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur avec session expirée en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer une session expirée manuellement
				expiredSessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, -1*time.Hour) // Expirée depuis 1 heure
				require.NoError(t, err)

				// Retourner les données de requête avec l'utilisateur pour le nettoyage et les headers
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
			CaseName: "Échec de déconnexion avec session déjà désactivée",
			CaseUrl:  "/auth/logout",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
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

				// Retourner les données de requête avec l'utilisateur pour le nettoyage et les headers
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
			// Préparer les données de requête
			requestData := testCase.RequestData()

			// Extraire les headers si présents
			var headers map[string]string
			if headersData, ok := requestData["_headers"]; ok {
				headers = headersData.(map[string]string)
				delete(requestData, "_headers") // Supprimer les headers des données de requête
			}

			// Créer le body de la requête JSON (vide pour logout)
			jsonBody, err := json.Marshal(requestData)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", testServer.URL+testCase.CaseUrl, bytes.NewBuffer(jsonBody))
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

// TestGetUserSessionsRoute teste la route GET de récupération des sessions utilisateur avec plusieurs cas
func TestGetUserSessionsRoute(t *testing.T) {
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
			CaseName: "Récupération réussie des sessions avec utilisateur authentifié",
			CaseUrl:  "/auth/sessions",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec plusieurs sessions actives en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer des sessions supplémentaires pour cet utilisateur
				_, _, _, err = testutils.CreateUserSession(user.User.UserID, 24*time.Hour)
				require.NoError(t, err)
				_, _, _, err = testutils.CreateUserSession(user.User.UserID, 12*time.Hour)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers nécessaires
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
			CaseName: "Récupération réussie des sessions avec utilisateur sans sessions",
			CaseUrl:  "/auth/sessions",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur sans session active en base
				user, err := testutils.GenerateAuthenticatedUser(false, true, false, false)
				require.NoError(t, err)

				// Créer une session active pour l'authentification
				sessionToken, _, _, err := testutils.CreateUserSession(user.User.UserID, 24*time.Hour)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur et les headers nécessaires
				return map[string]interface{}{
					"user": user, // Pour le nettoyage après le test
					"_headers": map[string]string{
						"Authorization": "Bearer " + sessionToken,
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
			CaseUrl:  "/auth/sessions",
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
			CaseUrl:  "/auth/sessions",
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
			ExpectedError:    common.ErrUserNotAuthenticated, // Le middleware retourne cette erreur pour un header vide
		},
		{
			CaseName: "Échec de récupération avec format de token invalide",
			CaseUrl:  "/auth/sessions",
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
			CaseUrl:  "/auth/sessions",
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
			CaseUrl:  "/auth/sessions",
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
			CaseUrl:  "/auth/sessions",
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

				// Vérifier que les données de réponse contiennent les informations attendues
				if response.Data != nil {
					// Convertir les données en slice pour vérification
					sessions, ok := response.Data.([]interface{})
					require.True(t, ok, "Les données de réponse devraient être un tableau de sessions")

					// Vérifier que chaque session a les champs attendus
					for _, sessionInterface := range sessions {
						session, ok := sessionInterface.(map[string]interface{})
						require.True(t, ok, "Chaque session devrait être un objet")

						// Vérifier la présence des champs attendus
						require.Contains(t, session, "user_session_id", "ID de session manquant")
						require.Contains(t, session, "user_id", "ID utilisateur manquant")
						require.Contains(t, session, "session_token", "Token de session manquant")
						require.Contains(t, session, "expires_at", "Date d'expiration manquante")
						require.Contains(t, session, "device_info", "Informations appareil manquantes")
						require.Contains(t, session, "ip_address", "Adresse IP manquante")
						require.Contains(t, session, "is_active", "Statut actif manquant")
						require.Contains(t, session, "created_at", "Date de création manquante")

						// Vérifier que les tokens sont masqués pour la sécurité
						sessionToken, ok := session["session_token"].(string)
						require.True(t, ok, "Le token de session devrait être une chaîne")
						require.Equal(t, "***", sessionToken, "Le token de session devrait être masqué")

						// Vérifier que le refresh token est nil (masqué)
						refreshToken := session["refresh_token"]
						require.Nil(t, refreshToken, "Le refresh token devrait être masqué (nil)")
					}
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

// TestDeleteSessionRoute teste la route DELETE de suppression d'une session spécifique avec plusieurs cas
func TestDeleteSessionRoute(t *testing.T) {
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
			CaseName: "Suppression réussie d'une session appartenant à l'utilisateur",
			CaseUrl:  "/auth/sessions/1", // L'ID sera remplacé dynamiquement
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec plusieurs sessions actives en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer des sessions supplémentaires pour cet utilisateur
				_, _, _, err = testutils.CreateUserSession(user.User.UserID, 24*time.Hour)
				require.NoError(t, err)
				sessionToken2, _, _, err := testutils.CreateUserSession(user.User.UserID, 12*time.Hour)
				require.NoError(t, err)

				// Récupérer l'ID de la session à supprimer
				var sessionID int
				err = common.DB.QueryRow("SELECT user_session_id FROM user_session WHERE session_token = ?", sessionToken2).Scan(&sessionID)
				require.NoError(t, err)

				// Retourner les données de préparation avec l'utilisateur, l'ID de session et les headers
				return map[string]interface{}{
					"user":      user, // Pour le nettoyage après le test
					"sessionID": sessionID,
					"_headers": map[string]string{
						"Authorization": "Bearer " + user.SessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessDeleteSession,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de suppression sans header Authorization",
			CaseUrl:  "/auth/sessions/1",
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
			CaseUrl:  "/auth/sessions/1",
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
			ExpectedError:    common.ErrUserNotAuthenticated, // Le middleware retourne cette erreur pour un header vide
		},
		{
			CaseName: "Échec de suppression avec format de token invalide",
			CaseUrl:  "/auth/sessions/1",
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
			CaseUrl:  "/auth/sessions/1",
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
			CaseUrl:  "/auth/sessions/1",
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
			CaseUrl:  "/auth/sessions/1",
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
		{
			CaseName: "Échec de suppression avec session_id manquant",
			CaseUrl:  "/auth/sessions",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				user.User.UserID = 0

				// Retourner les données de préparation avec l'utilisateur et les headers
				return map[string]interface{}{
					"user":      user, // Pour le nettoyage après le test
					"sessionID": "",
					"_headers": map[string]string{
						"Authorization": "Bearer " + user.SessionToken,
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de suppression avec session_id inexistant",
			CaseUrl:  "/auth/sessions/99999", // ID qui n'existe pas
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec session active en base
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
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionNotFound,
		},
		{
			CaseName: "Échec de suppression avec session appartenant à un autre utilisateur",
			CaseUrl:  "/auth/sessions/1", // L'ID sera remplacé dynamiquement
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer deux utilisateurs avec sessions actives
				user1, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				user2, err := testutils.GenerateAuthenticatedUser(true, true, false, false)
				require.NoError(t, err)

				// Créer une session pour le deuxième utilisateur
				sessionToken2, _, _, err := testutils.CreateUserSession(user2.User.UserID, 24*time.Hour)
				require.NoError(t, err)

				// Récupérer l'ID de la session du deuxième utilisateur
				var sessionID int
				err = common.DB.QueryRow("SELECT user_session_id FROM user_session WHERE session_token = ?", sessionToken2).Scan(&sessionID)
				require.NoError(t, err)

				// Retourner les données de préparation avec les utilisateurs, l'ID de session et les headers
				return map[string]interface{}{
					"user1":     user1, // Utilisateur authentifié
					"user2":     user2, // Propriétaire de la session
					"sessionID": sessionID,
					"_headers": map[string]string{
						"Authorization": "Bearer " + user1.SessionToken, // user1 essaie de supprimer la session de user2
						"Content-Type":  "application/json",
					},
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionNotFound,
		},
		{
			CaseName: "Échec de suppression avec session_id invalide (non numérique)",
			CaseUrl:  "/auth/sessions/invalid",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST GET/DELETE
				// Créer un utilisateur avec session active en base
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
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionNotFound,
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

			// Construire l'URL avec le session_id si fourni
			url := testCase.CaseUrl
			if sessionID, ok := setupData["sessionID"].(int); ok {
				url = fmt.Sprintf("/auth/sessions/%d", sessionID)
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

			// Vérifier la réponse selon le cas de test
			if testCase.ExpectedHttpCode != http.StatusNotFound {
				// Lire et parser la réponse JSON
				var response common.JSONResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoErrorf(t, err, "Erreur lors du parsing de la réponse JSON: %", resp.Body)

				if testCase.ExpectedHttpCode == http.StatusOK {
					require.True(t, response.Success, "La réponse devrait être un succès")
					require.Equal(t, testCase.ExpectedMessage, response.Message, "Message de succès incorrect")
					require.Empty(t, response.Error, "Pas d'erreur attendue")
				} else {
					require.False(t, response.Success, "La réponse devrait être un échec")
					require.Equal(t, testCase.ExpectedError, response.Error, "Message d'erreur incorrect")
					require.Empty(t, response.Message, "Pas de message de succès attendu")
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}
