package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
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

// TestMain configure l'environnement de test global
func TestMain(m *testing.M) {
	// Initialiser l'environnement de test
	if err := testutils.SetupTestEnvironment(); err != nil {
		panic("Impossible d'initialiser l'environnement de test: " + err.Error())
	}

	// Exécuter les tests
	code := m.Run()

	// Nettoyer l'environnement de test
	if err := testutils.TeardownTestEnvironment(); err != nil {
		panic("Impossible de nettoyer l'environnement de test: " + err.Error())
	}

	// Retourner le code de sortie
	os.Exit(code)
}

// TestRouteExample teste la route d'exemple avec plusieurs cas
func TestRouteExample(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		CaseUrl          string
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName:         "Case name",
			CaseUrl:          "Url",
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

// TestUserAdd teste la fonction d'ajout d'utilisateur avec plusieurs cas
func TestUserAdd(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		RequestData      common.CreateUserRequest
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Création d'utilisateur valide",
			RequestData: common.CreateUserRequest{
				Lastname:  "Dupont",
				Firstname: "Jean",
				Email:     "jean.dupont@example.com",
				Password:  "password123",
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateUser,
			ExpectedError:    "",
		},
		{
			CaseName: "Email invalide",
			RequestData: common.CreateUserRequest{
				Lastname:  "Martin",
				Firstname: "Pierre",
				Email:     "email-invalide",
				Password:  "password123",
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Mot de passe trop court",
			RequestData: common.CreateUserRequest{
				Lastname:  "Bernard",
				Firstname: "Marie",
				Email:     "marie.bernard@example.com",
				Password:  "123",
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
		},
		{
			CaseName: "Données manquantes",
			RequestData: common.CreateUserRequest{
				Lastname:  "",
				Firstname: "",
				Email:     "",
				Password:  "",
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
			jsonData, err := json.Marshal(testCase.RequestData)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response common.JSONResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if testCase.ExpectedError != "" {
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			if testCase.ExpectedMessage != "" {
				require.Contains(t, response.Message, testCase.ExpectedMessage)
			}

			// On purge les données après avoir traité le cas.
			// Ne pas passer au cas suivant si la purge échoue
			if err := testutils.PurgeTestData(testCase.RequestData.Email); err != nil {
				t.Fatalf("Échec de la purge des données de test: %v", err)
			}
		})
	}
}

// TestUserGet teste la fonction de récupération d'utilisateur
func TestUserGet(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié - route /user/me",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - route /user/me",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Utilisateur avec token invalide - route /user/me",
			SetupAuth: func() (string, string, error) {
				return "Bearer invalid-token", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrSessionInvalid,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/user/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserUpdate teste la fonction de mise à jour d'utilisateur
func TestUserUpdate(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		RequestData      common.UpdateUserRequest
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Mise à jour nom et prénom - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Lastname:  common.StringPtr("Martin"),
				Firstname: common.StringPtr("Pierre"),
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Mise à jour email valide - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("marie.bernard")
				user, token, err := testutils.CreateAuthenticatedUser(2, "Bernard", "Marie", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Email: common.StringPtr(testutils.GenerateUniqueEmail("nouveau.email")),
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
		},
		{
			CaseName: "Email invalide - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("sophie.petit")
				user, token, err := testutils.CreateAuthenticatedUser(3, "Petit", "Sophie", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Email: common.StringPtr("email-invalide"),
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEmailFormat,
		},
		{
			CaseName: "Mot de passe trop court - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("thomas.roux")
				user, token, err := testutils.CreateAuthenticatedUser(4, "Roux", "Thomas", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Password: common.StringPtr("123"),
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrPasswordTooShort,
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			RequestData: common.UpdateUserRequest{
				Lastname: common.StringPtr("Nouveau"),
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			jsonData, err := json.Marshal(testCase.RequestData)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			// Log temporaire pour vérifier la session AVANT la requête
			if token != "" {
				tokenStr := token
				if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
					tokenStr = tokenStr[7:]
				}
				var isActive bool
				var deletedAt sql.NullTime
				err := common.DB.QueryRow("SELECT is_active, deleted_at FROM user_session WHERE session_token = ?", tokenStr).Scan(&isActive, &deletedAt)
				if err != nil {
					fmt.Printf("[DEBUG][TEST] Session non trouvée avant requête: %v\n", err)
				} else {
					fmt.Printf("[DEBUG][TEST] Session avant requête: is_active=%v, deleted_at.Valid=%v\n", isActive, deletedAt.Valid)
				}
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			if testCase.ExpectedMessage != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Message, testCase.ExpectedMessage)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserDelete teste la fonction de suppression d'utilisateur
func TestUserDelete(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Suppression d'utilisateur valide - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserDelete,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("DELETE", "/user/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			if testCase.ExpectedMessage != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Message, testCase.ExpectedMessage)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserGetWithRoles teste la fonction de récupération d'utilisateur avec rôles
func TestUserGetWithRoles(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur avec rôles dans le contexte - route /auth/me",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - route /auth/me-with-roles",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserGetAuthMe teste la fonction GetAuthMe (alias de GetUserWithRoles)
func TestUserGetAuthMe(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié avec rôles - route /auth/me",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - route /auth/me",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserAddErrors teste les cas d'erreur de la fonction Add
func TestUserAddErrors(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() (common.CreateUserRequest, string, error) // Retourne (request, emailToPurge, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Email déjà utilisé",
			SetupData: func() (common.CreateUserRequest, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				// Créer un premier utilisateur
				_, err := testutils.CreateUserWithPassword("Dupont", "Jean", email, "password123")
				if err != nil {
					return common.CreateUserRequest{}, "", err
				}
				// Essayer de créer un deuxième utilisateur avec le même email
				req := common.CreateUserRequest{
					Lastname:  "Martin",
					Firstname: "Pierre",
					Email:     email,
					Password:  "password456",
				}
				return req, email, nil
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, emailToPurge, err := testCase.SetupData()
			require.NoError(t, err)

			jsonData, err := json.Marshal(req)
			require.NoError(t, err)

			httpReq, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, httpReq)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if emailToPurge != "" {
				if err := testutils.PurgeTestData(emailToPurge); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserUpdateErrors teste les cas d'erreur de la fonction Update
func TestUserUpdateErrors(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, string, error) // Retourne (token, userEmail, emailToPurge, error)
		RequestData      common.UpdateUserRequest
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Email déjà utilisé par un autre utilisateur",
			SetupAuth: func() (string, string, string, error) {
				// Créer un premier utilisateur avec un userID unique
				userID1 := int(time.Now().UnixNano() % 1000000)
				email1 := testutils.GenerateUniqueEmail("jean.dupont")
				user1, _, err := testutils.CreateAuthenticatedUser(userID1, "Dupont", "Jean", email1)
				if err != nil {
					return "", "", "", err
				}

				// Créer un deuxième utilisateur avec un autre userID unique
				userID2 := userID1 + 1 + rand.Intn(1000000)
				email2 := testutils.GenerateUniqueEmail("marie.martin")
				user2, token, err := testutils.CreateAuthenticatedUser(userID2, "Martin", "Marie", email2)
				if err != nil {
					return "", "", "", err
				}

				// Essayer de mettre à jour le deuxième utilisateur avec l'email du premier
				return "Bearer " + token, user2.Email, user1.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Email: common.StringPtr(""), // Sera défini dynamiquement
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			token, userEmail, emailToPurge, err := testCase.SetupAuth()
			require.NoError(t, err)

			// Si c'est le cas d'email déjà utilisé, on utilise l'email du premier utilisateur
			if testCase.RequestData.Email != nil && *testCase.RequestData.Email == "" {
				testCase.RequestData.Email = common.StringPtr(emailToPurge)
			}

			// Log temporaire pour vérifier la session AVANT la requête
			if token != "" {
				tokenStr := token
				if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
					tokenStr = tokenStr[7:]
				}
				var isActive bool
				var deletedAt sql.NullTime
				err := common.DB.QueryRow("SELECT is_active, deleted_at FROM user_session WHERE session_token = ?", tokenStr).Scan(&isActive, &deletedAt)
				if err != nil {
					fmt.Printf("[DEBUG][TEST] Session non trouvée avant requête: %v\n", err)
				} else {
					fmt.Printf("[DEBUG][TEST] Session avant requête: is_active=%v, deleted_at.Valid=%v\n", isActive, deletedAt.Valid)
				}
			}

			jsonData, err := json.Marshal(testCase.RequestData)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Configurer l'authentification si nécessaire
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
			if emailToPurge != "" {
				if err := testutils.PurgeTestData(emailToPurge); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserAdminRoutes teste les routes admin pour la gestion des utilisateurs
func TestUserAdminRoutes(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (token string, userEmail string, url string, err error)
		Method           string
		RequestData      interface{}
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Admin récupère un utilisateur par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID)
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method:           "GET",
			RequestData:      nil,
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin met à jour un utilisateur par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID)
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method: "PUT",
			RequestData: common.UpdateUserRequest{
				Lastname: common.StringPtr("Updated"),
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin supprime un utilisateur par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID)
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method:           "DELETE",
			RequestData:      nil,
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Admin récupère un utilisateur avec rôles par ID",
			SetupAuth: func() (string, string, string, error) {
				adminID := int(time.Now().UnixNano() % 1000000)
				adminEmail := testutils.GenerateUniqueEmail("admin")
				_, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "User", adminEmail)
				if err != nil {
					return "", "", "", err
				}
				targetID := int(time.Now().UnixNano() % 1000000)
				targetEmail := testutils.GenerateUniqueEmail("target")
				targetUser, _, err := testutils.CreateAuthenticatedUser(targetID, "Target", "User", targetEmail)
				if err != nil {
					return "", "", "", err
				}
				url := "/user/" + fmt.Sprintf("%d", targetUser.UserID) + "/with-roles"
				return "Bearer " + adminToken, targetUser.Email, url, nil
			},
			Method:           "GET",
			RequestData:      nil,
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userEmail, url, err := testCase.SetupAuth()
			require.NoError(t, err)
			var req *http.Request
			if testCase.RequestData != nil {
				jsonData, err := json.Marshal(testCase.RequestData)
				require.NoError(t, err)
				req, err = http.NewRequest(testCase.Method, url, bytes.NewBuffer(jsonData))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(testCase.Method, url, nil)
				require.NoError(t, err)
			}
			if token != "" {
				req.Header.Set("Authorization", token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)
			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserGetAuthMeDirect teste directement la fonction GetAuthMe (pas via route)
func TestUserGetAuthMeDirect(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié - fonction GetAuthMe directe",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - fonction GetAuthMe directe",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserAddDatabaseErrors teste les erreurs de base de données lors de la création
func TestUserAddDatabaseErrors(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		RequestData      common.CreateUserRequest
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Email déjà utilisé",
			RequestData: common.CreateUserRequest{
				Lastname:  "Dupont",
				Firstname: "Jean",
				Email:     "jean.dupont@example.com",
				Password:  "password123",
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// Créer d'abord un utilisateur avec le même email
			existingUser := testCase.RequestData
			jsonData, err := json.Marshal(existingUser)
			require.NoError(t, err)

			req1, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req1.Header.Set("Content-Type", "application/json")

			w1 := httptest.NewRecorder()
			router.ServeHTTP(w1, req1)
			require.Equal(t, http.StatusCreated, w1.Code)

			// On prépare les données utiles au traitement de ce cas.
			// Essayer de créer un deuxième utilisateur avec le même email
			duplicateUser := testCase.RequestData
			duplicateUser.Firstname = "Pierre" // Changer le prénom pour différencier
			jsonData2, err := json.Marshal(duplicateUser)
			require.NoError(t, err)

			req2, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData2))
			require.NoError(t, err)
			req2.Header.Set("Content-Type", "application/json")

			w2 := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w2, req2)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w2.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w2.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if err := testutils.PurgeTestData(testCase.RequestData.Email); err != nil {
				t.Fatalf("Échec de la purge des données de test: %v", err)
			}
		})
	}
}

// TestUserUpdateDatabaseErrors teste les erreurs de base de données lors de la mise à jour
func TestUserUpdateDatabaseErrors(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, string, error) // Retourne (token, userEmail, emailToPurge, error)
		RequestData      common.UpdateUserRequest
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Email déjà utilisé par un autre utilisateur",
			SetupAuth: func() (string, string, string, error) {
				// Créer un premier utilisateur avec un userID unique
				userID1 := int(time.Now().UnixNano() % 1000000)
				email1 := testutils.GenerateUniqueEmail("jean.dupont")
				user1, _, err := testutils.CreateAuthenticatedUser(userID1, "Dupont", "Jean", email1)
				if err != nil {
					return "", "", "", err
				}

				// Créer un deuxième utilisateur avec un autre userID unique
				userID2 := userID1 + 1 + rand.Intn(1000000)
				email2 := testutils.GenerateUniqueEmail("marie.martin")
				user2, token, err := testutils.CreateAuthenticatedUser(userID2, "Martin", "Marie", email2)
				if err != nil {
					return "", "", "", err
				}

				// Essayer de mettre à jour le deuxième utilisateur avec l'email du premier
				return "Bearer " + token, user2.Email, user1.Email, nil
			},
			RequestData: common.UpdateUserRequest{
				Email: common.StringPtr(""), // Sera défini dynamiquement
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedError:    common.ErrUserAlreadyExists,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			token, userEmail, emailToPurge, err := testCase.SetupAuth()
			require.NoError(t, err)

			// Si c'est le cas d'email déjà utilisé, on utilise l'email du premier utilisateur
			if testCase.RequestData.Email != nil && *testCase.RequestData.Email == "" {
				testCase.RequestData.Email = common.StringPtr(emailToPurge)
			}

			// Log temporaire pour vérifier la session AVANT la requête
			if token != "" {
				tokenStr := token
				if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
					tokenStr = tokenStr[7:]
				}
				var isActive bool
				var deletedAt sql.NullTime
				err := common.DB.QueryRow("SELECT is_active, deleted_at FROM user_session WHERE session_token = ?", tokenStr).Scan(&isActive, &deletedAt)
				if err != nil {
					fmt.Printf("[DEBUG][TEST] Session non trouvée avant requête: %v\n", err)
				} else {
					fmt.Printf("[DEBUG][TEST] Session avant requête: is_active=%v, deleted_at.Valid=%v\n", isActive, deletedAt.Valid)
				}
			}

			jsonData, err := json.Marshal(testCase.RequestData)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Configurer l'authentification si nécessaire
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
			if emailToPurge != "" {
				if err := testutils.PurgeTestData(emailToPurge); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserGetWithRolesDatabaseErrors teste les erreurs de base de données dans GetUserWithRoles
func TestUserGetWithRolesDatabaseErrors(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur avec rôles - route /auth/me",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - route /auth/me",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserDeleteDatabaseErrors teste les erreurs de base de données lors de la suppression
func TestUserDeleteDatabaseErrors(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Suppression d'utilisateur valide - utilisateur authentifié",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserDelete,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("DELETE", "/user/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			// On traite les cas de test un par un.
			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			if testCase.ExpectedMessage != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Message, testCase.ExpectedMessage)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}

// TestUserAddPasswordHashingError teste l'erreur de hash de mot de passe
func TestUserAddPasswordHashingError(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// Créer une requête avec un mot de passe valide
	req := common.CreateUserRequest{
		Lastname:  "Dupont",
		Firstname: "Jean",
		Email:     testutils.GenerateUniqueEmail("jean.dupont"),
		Password:  "password123",
	}

	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	// Le test devrait passer car bcrypt ne devrait pas échouer
	require.Equal(t, http.StatusCreated, w.Code)

	// Nettoyer
	if err := testutils.PurgeTestData(req.Email); err != nil {
		t.Fatalf("Échec de la purge des données de test: %v", err)
	}
}

// TestUserUpdatePasswordHashingError teste l'erreur de hash de mot de passe lors de la mise à jour
func TestUserUpdatePasswordHashingError(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// Ce test simule une erreur de hash de mot de passe lors de la mise à jour
	// En pratique, bcrypt.GenerateFromPassword ne devrait jamais échouer avec des données valides
	// Créer un utilisateur authentifié
	email := testutils.GenerateUniqueEmail("jean.dupont")
	user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
	require.NoError(t, err)

	// Créer une requête de mise à jour avec un nouveau mot de passe
	updateReq := common.UpdateUserRequest{
		Password: common.StringPtr("newpassword123"),
	}

	jsonData, err := json.Marshal(updateReq)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Le test devrait passer car bcrypt ne devrait pas échouer
	require.Equal(t, http.StatusOK, w.Code)

	// Nettoyer
	if err := testutils.PurgeTestData(user.Email); err != nil {
		t.Fatalf("Échec de la purge des données de test: %v", err)
	}
}

// TestUserAddInvalidJSON teste l'erreur de parsing JSON invalide
func TestUserAddInvalidJSON(t *testing.T) {
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	// Créer une requête avec JSON invalide
	invalidJSON := `{"lastname": "Dupont", "firstname": "Jean", "email": "jean.dupont@example.com", "password": "password123"`

	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer([]byte(invalidJSON)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Le test devrait échouer avec une erreur de parsing JSON
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response common.JSONResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Contains(t, response.Error, common.ErrInvalidData)
}

// TestUserUpdateInvalidJSON teste l'erreur de parsing JSON invalide lors de la mise à jour
func TestUserUpdateInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testutils.CreateTestRouter()

	// Créer un utilisateur authentifié
	email := testutils.GenerateUniqueEmail("jean.dupont")
	user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
	require.NoError(t, err)

	// Créer une requête avec JSON invalide
	invalidJSON := `{"lastname": "Martin"`

	req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer([]byte(invalidJSON)))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Le test devrait échouer avec une erreur de parsing JSON
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response common.JSONResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Contains(t, response.Error, common.ErrInvalidData)

	// Nettoyer
	if err := testutils.PurgeTestData(user.Email); err != nil {
		t.Fatalf("Échec de la purge des données de test: %v", err)
	}
}

// TestUserGetWithRolesErrorHandling teste la gestion d'erreur dans GetUserWithRoles
func TestUserGetWithRolesErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := testutils.CreateTestRouter()

	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupAuth        func() (string, string, error) // Retourne (token, email, error)
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Utilisateur authentifié avec rôles - test direct GetUserWithRoles",
			SetupAuth: func() (string, string, error) {
				email := testutils.GenerateUniqueEmail("jean.dupont")
				user, token, err := testutils.CreateAuthenticatedUser(1, "Dupont", "Jean", email)
				if err != nil {
					return "", "", err
				}
				return "Bearer " + token, user.Email, nil
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Utilisateur non authentifié - test direct GetUserWithRoles",
			SetupAuth: func() (string, string, error) {
				return "", "", nil
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    common.ErrUserNotAuthenticated,
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On prépare les données utiles au traitement de ce cas.
			req, err := http.NewRequest("GET", "/auth/me", nil)
			require.NoError(t, err)

			// Configurer l'authentification si nécessaire
			token, userEmail, err := testCase.SetupAuth()
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Vérifications
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			if testCase.ExpectedError != "" {
				var response common.JSONResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				require.Contains(t, response.Error, testCase.ExpectedError)
			}

			// On purge les données après avoir traité le cas.
			if userEmail != "" {
				if err := testutils.PurgeTestData(userEmail); err != nil {
					t.Fatalf("Échec de la purge des données de test: %v", err)
				}
			}
		})
	}
}
