package role_test

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

// TestListRolesRoute teste la route GET de récupération de la liste des rôles avec plusieurs cas
func TestListRolesRoute(t *testing.T) {
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
			CaseName: "Récupération réussie de la liste des rôles avec rôles existants",
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "User", "Rôle utilisateur standard")
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Moderator", "Rôle modérateur")
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer des rôles sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "Admin", "Rôle administrateur")
				require.NoError(t, err)

				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer des rôles

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
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer des rôles

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
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin avec session expirée en base
				admin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(admin.User.UserID, -1*time.Hour)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":        admin,
					"sessionToken": expiredSessionToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, admin.SessionToken)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération par un utilisateur non admin",
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur normal (non admin) authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Récupération réussie avec rôles supprimés (soft delete)",
			CaseUrl:  "/roles",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)

				_, err = common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "User", "Rôle utilisateur standard")
				require.NoError(t, err)

				// Supprimer un rôle (soft delete)
				_, err = common.DB.Exec(`
					UPDATE roles 
					SET deleted_at = NOW() 
					WHERE name = 'User'
				`)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupData := testCase.SetupData()

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+testCase.CaseUrl, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header d'authentification si disponible
			if admin, ok := setupData["admin"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+admin.SessionToken)
			} else if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
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

				// Vérifier que les données sont présentes
				if response.Data != nil {
					// Vérifier que les données sont un tableau de rôles
					rolesData, ok := response.Data.([]interface{})
					require.True(t, ok, "Les données devraient être un tableau de rôles")

					// Pour les cas avec rôles, vérifier qu'il y en a
					if strings.Contains(testCase.CaseName, "avec rôles existants") {
						require.Greater(t, len(rolesData), 0, "Il devrait y avoir au moins un rôle")
					}

					// Pour les cas sans rôles, vérifier qu'il n'y en a pas
					if strings.Contains(testCase.CaseName, "sans rôles") {
						require.Equal(t, 0, len(rolesData), "Il ne devrait y avoir aucun rôle")
					}

					// Pour les cas avec rôles supprimés, vérifier qu'il n'y a que les rôles actifs
					if strings.Contains(testCase.CaseName, "supprimés") {
						require.Equal(t, 1, len(rolesData), "Il devrait y avoir un seul rôle actif")
					}
				} else {
					// Pour les cas sans rôles, les données peuvent être nulles
					if strings.Contains(testCase.CaseName, "sans rôles") {
						// C'est normal que les données soient nulles pour un tableau vide
					} else {
						require.NotNil(t, response.Data, "Les données de réponse ne devraient pas être nulles")
					}
				}
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestGetRoleRoute teste la route GET de récupération d'un rôle par ID avec plusieurs cas
func TestGetRoleRoute(t *testing.T) {
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
			CaseName: "Récupération réussie d'un rôle par son ID",
			CaseUrl:  "/roles/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)

				// Créer un rôle de test
				result, err := common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
				require.NoError(t, err)
				roleID, err := result.LastInsertId()
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":  admin,
					"roleID": roleID,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de récupération sans header Authorization",
			CaseUrl:  "/roles/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un rôle sans utilisateur authentifié
				_, err := common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
				require.NoError(t, err)

				return map[string]interface{}{}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de récupération avec header Authorization vide",
			CaseUrl:  "/roles/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un rôle
				_, err := common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
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
			CaseUrl:  "/roles/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un rôle
				_, err := common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
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
			CaseUrl:  "/roles/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin avec session expirée en base
				admin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(admin.User.UserID, -1*time.Hour)
				require.NoError(t, err)

				// Créer un rôle
				_, err = common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":        admin,
					"sessionToken": expiredSessionToken,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération avec session désactivée",
			CaseUrl:  "/roles/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, admin.SessionToken)
				require.NoError(t, err)

				// Créer un rôle
				_, err = common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de récupération par un utilisateur non admin",
			CaseUrl:  "/roles/1",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur normal (non admin) authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)

				// Créer un rôle
				_, err = common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
				require.NoError(t, err)

				return map[string]interface{}{
					"user": user,
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de récupération avec role_id inexistant",
			CaseUrl:  "/roles/99999",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrRoleNotFound,
		},
		{
			CaseName: "Échec de récupération avec role_id invalide (non numérique)",
			CaseUrl:  "/roles/invalid",
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrRoleNotFound,
		},
		{
			CaseName: "Échec de récupération d'un rôle supprimé",
			CaseUrl:  "/roles/1", // Sera remplacé par l'ID réel
			SetupData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION D'UN APPEL GET/DELETE
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)

				// Créer un rôle
				result, err := common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "TestRole", "Rôle de test")
				require.NoError(t, err)
				roleID, err := result.LastInsertId()
				require.NoError(t, err)

				// Supprimer le rôle (soft delete)
				_, err = common.DB.Exec(`
					UPDATE roles 
					SET deleted_at = NOW() 
					WHERE role_id = ?
				`, roleID)
				require.NoError(t, err)

				return map[string]interface{}{
					"admin":  admin,
					"roleID": roleID,
				}
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrRoleNotFound,
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
			if roleID, ok := setupData["roleID"].(int64); ok {
				url = "/roles/" + strconv.FormatInt(roleID, 10)
			}

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", testServer.URL+url, nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le header d'authentification si disponible
			if admin, ok := setupData["admin"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+admin.SessionToken)
			} else if user, ok := setupData["user"].(*testutils.AuthenticatedUser); ok {
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

				// Vérifier que les données du rôle sont présentes
				roleData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet rôle")
				require.Contains(t, roleData, "role_id", "Le rôle devrait avoir un ID")
				require.Contains(t, roleData, "name", "Le rôle devrait avoir un nom")
				require.Contains(t, roleData, "description", "Le rôle devrait avoir une description")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}

// TestCreateRoleRoute teste la route POST de création d'un rôle avec plusieurs cas
func TestCreateRoleRoute(t *testing.T) {
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
			CaseName: "Création réussie d'un rôle avec toutes les données",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
					"requestBody": map[string]interface{}{
						"name":        "Moderator",
						"description": "Rôle modérateur avec permissions limitées",
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateRole,
			ExpectedError:    "",
		},
		{
			CaseName: "Création réussie d'un rôle avec description vide",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
					"requestBody": map[string]interface{}{
						"name":        "Editor",
						"description": "",
					},
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateRole,
			ExpectedError:    "",
		},
		{
			CaseName: "Échec de création sans header Authorization",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"requestBody": map[string]interface{}{
						"name":        "TestRole",
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de création avec header Authorization vide",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"authHeader": "",
					"requestBody": map[string]interface{}{
						"name":        "TestRole",
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Échec de création avec token invalide",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				return map[string]interface{}{
					"authHeader": "Bearer invalid_token_12345",
					"requestBody": map[string]interface{}{
						"name":        "TestRole",
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec session expirée",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin avec session expirée en base
				admin, err := testutils.GenerateAuthenticatedAdmin(false, true, true, true)
				require.NoError(t, err)
				expiredSessionToken, _, _, err := testutils.CreateUserSession(admin.User.UserID, -1*time.Hour)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin":        admin,
					"sessionToken": expiredSessionToken,
					"requestBody": map[string]interface{}{
						"name":        "TestRole",
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création avec session désactivée",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				// Désactiver la session
				_, err = common.DB.Exec(`
					UPDATE user_session 
					SET is_active = FALSE 
					WHERE session_token = ?
				`, admin.SessionToken)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
					"requestBody": map[string]interface{}{
						"name":        "TestRole",
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrSessionInvalid,
		},
		{
			CaseName: "Échec de création par un utilisateur non admin",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur normal (non admin) authentifié avec session active en base
				user, err := testutils.GenerateAuthenticatedUser(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"user": user,
					"requestBody": map[string]interface{}{
						"name":        "TestRole",
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInsufficientPermissions,
		},
		{
			CaseName: "Échec de création avec nom manquant",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
					"requestBody": map[string]interface{}{
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec nom vide",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin": admin,
					"requestBody": map[string]interface{}{
						"name":        "",
						"description": "Rôle de test",
					},
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Échec de création avec nom déjà existant",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)

				// Créer un rôle avec le même nom
				_, err = common.DB.Exec(`
					INSERT INTO roles (name, description, created_at) 
					VALUES (?, ?, NOW())
				`, "DuplicateRole", "Rôle existant")
				require.NoError(t, err)

				return map[string]interface{}{
					"admin": admin,
					"requestBody": map[string]interface{}{
						"name":        "DuplicateRole",
						"description": "Rôle en doublon",
					},
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrRoleAlreadyExists,
		},
		{
			CaseName: "Échec de création avec données JSON vides",
			CaseUrl:  "/roles",
			RequestData: func() map[string]interface{} {
				// DOIT CONTENIR L'ENSEMBLE DES INSTRUCTIONS QUI PREPARENT LE CAS A LA RECEPTION DE LA REQUEST POST/PUT
				// Créer un utilisateur admin authentifié avec session active en base
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true, true, true)
				require.NoError(t, err)
				return map[string]interface{}{
					"admin":       admin,
					"requestBody": map[string]interface{}{},
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
			requestData := testCase.RequestData()

			// Extraire les données de requête
			requestBody, ok := requestData["requestBody"].(map[string]interface{})
			require.True(t, ok, "Le corps de la requête doit être présent")

			// Préparer la requête JSON
			jsonData, err := json.Marshal(requestBody)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", testServer.URL+testCase.CaseUrl, bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")
			req.Header.Set("Content-Type", "application/json")

			// Ajouter le header d'authentification si disponible
			if admin, ok := requestData["admin"].(*testutils.AuthenticatedUser); ok {
				req.Header.Set("Authorization", "Bearer "+admin.SessionToken)
			} else if user, ok := requestData["user"].(*testutils.AuthenticatedUser); ok {
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

				// Vérifier que les données du rôle créé sont présentes
				roleData, ok := response.Data.(map[string]interface{})
				require.True(t, ok, "Les données devraient être un objet rôle")
				require.Contains(t, roleData, "role_id", "Le rôle devrait avoir un ID")
			}

			// On purge les données après avoir traité le cas.
			testutils.PurgeAllTestUsers()
		})
	}
}
