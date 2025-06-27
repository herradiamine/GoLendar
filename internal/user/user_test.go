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

	"github.com/stretchr/testify/require"
)

// getStringValue retourne la valeur d'un pointeur string ou "<nil>" si nil
func getStringValue(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

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

// TestCreateUser teste la route POST /user (création d'un nouvel utilisateur)
func TestCreateUser(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		RequestData      func() any
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		SetupData        func() string
		CleanupData      func()
	}{
		{
			CaseName: "Création d'utilisateur avec données valides",
			RequestData: func() any {
				nonExistingUser, err := testutils.GenerateAuthenticatedUser(false, false)
				if err != nil {
					// En cas d'erreur, utiliser un email par défaut pour le test
					t.Errorf("Erreur: %v", err)
					return nil
				}
				return common.CreateUserRequest{
					Lastname:  nonExistingUser.User.Lastname,
					Firstname: nonExistingUser.User.Firstname,
					Email:     nonExistingUser.User.Email,
					Password:  nonExistingUser.Password,
				}
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  common.MsgSuccessCreateUser,
			ExpectedError:    "",
			SetupData:        func() string { return "Données de test préparées" },
			CleanupData:      func() { /* Nettoyage automatique par la base de test */ },
		},
		{
			CaseName: "Création d'utilisateur avec email déjà existant",
			RequestData: func() any {
				existingUser, err := testutils.GenerateAuthenticatedUser(false, true)
				if err != nil {
					// En cas d'erreur, utiliser un email par défaut pour le test
					t.Errorf("Erreur: %v", err)
					return nil
				}
				// L'email sera défini après le setup
				return common.CreateUserRequest{
					Lastname:  "Martin",
					Firstname: "Marie",
					Email:     existingUser.User.Email, // Utiliser l'email de l'utilisateur créé
					Password:  "MotDePasse456!",
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserAlreadyExists,
			SetupData:        func() string { return "" },
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Création d'utilisateur avec email invalide",
			RequestData: func() any {
				return common.CreateUserRequest{
					Lastname:  "Durand",
					Firstname: "Pierre",
					Email:     "email-invalide",
					Password:  "MotDePasse789!",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
			SetupData:        func() string { return "Validation email" },
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Création d'utilisateur avec mot de passe trop court",
			RequestData: func() any {
				return common.CreateUserRequest{
					Lastname:  "Leroy",
					Firstname: "Sophie",
					Email:     "sophie.leroy@example.com",
					Password:  "123",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
			SetupData:        func() string { return "Validation mot de passe" },
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Création d'utilisateur avec nom manquant",
			RequestData: func() any {
				return common.CreateUserRequest{
					Lastname:  "",
					Firstname: "Paul",
					Email:     "paul@example.com",
					Password:  "MotDePasse123!",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
			SetupData:        func() string { return "Validation nom manquant" },
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Création d'utilisateur avec prénom manquant",
			RequestData: func() any {
				return common.CreateUserRequest{
					Lastname:  "Moreau",
					Firstname: "",
					Email:     "moreau@example.com",
					Password:  "MotDePasse123!",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
			SetupData:        func() string { return "Validation prénom manquant" },
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Création d'utilisateur avec email manquant",
			RequestData: func() any {
				return common.CreateUserRequest{
					Lastname:  "Petit",
					Firstname: "Lucie",
					Email:     "",
					Password:  "MotDePasse123!",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
			SetupData:        func() string { return "Validation email manquant" },
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Création d'utilisateur avec mot de passe manquant",
			RequestData: func() any {
				return common.CreateUserRequest{
					Lastname:  "Roux",
					Firstname: "Thomas",
					Email:     "thomas.roux@example.com",
					Password:  "",
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidData,
			SetupData:        func() string { return "Validation mot de passe manquant" },
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			setupInfo := testCase.SetupData()
			t.Logf("Données de setup: %s", setupInfo)

			// Obtenir les données de requête (peut dépendre du setup)
			requestData := testCase.RequestData()

			// Créer le routeur de test
			router := testutils.CreateTestRouter()

			// Préparer la requête JSON
			jsonData, err := json.Marshal(requestData)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			req.Header.Set("Content-Type", "application/json")

			// Créer le recorder pour capturer la réponse
			w := httptest.NewRecorder()

			// Exécuter la requête
			router.ServeHTTP(w, req)

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, w.Code,
				"Code HTTP attendu %d, obtenu %d", testCase.ExpectedHttpCode, w.Code)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message,
					"Message attendu '%s', obtenu '%s'", testCase.ExpectedMessage, response.Message)
			}

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error,
					"Erreur attendue '%s', obtenue '%s'", testCase.ExpectedError, response.Error)
			}

			// Vérifications spécifiques pour la création réussie
			if testCase.ExpectedHttpCode == http.StatusCreated {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")
				require.NotNil(t, response.Data, "La réponse devrait contenir des données")

				// Vérifier que l'ID utilisateur est retourné
				if data, ok := response.Data.(map[string]interface{}); ok {
					require.Contains(t, data, "user_id", "La réponse devrait contenir un user_id")
					require.NotNil(t, data["user_id"], "Le user_id ne devrait pas être null")
				}
			}

			// On purge les données après avoir traité le cas.
			testCase.CleanupData()
		})
	}
}

// TestGetUserMe teste la route GET /user/me (récupération des données de l'utilisateur connecté)
func TestGetUserMe(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() *testutils.AuthenticatedUser
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		CleanupData      func()
	}{
		{
			CaseName: "Récupération des données avec utilisateur normal authentifié",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					t.Errorf("Erreur: %v", err)
					return nil
				}
				return user
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Récupération des données avec utilisateur admin authentifié",
			SetupData: func() *testutils.AuthenticatedUser {
				admin, err := testutils.GenerateAuthenticatedAdmin(true, true)
				if err != nil {
					t.Errorf("Erreur: %v", err)
					return nil
				}
				return admin
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "",
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Tentative d'accès sans authentification",
			SetupData: func() *testutils.AuthenticatedUser {
				return nil // Pas d'utilisateur authentifié
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			authenticatedUser := testCase.SetupData()

			if authenticatedUser != nil {
				t.Logf("Utilisateur de test: %s %s (%s)",
					authenticatedUser.User.Firstname,
					authenticatedUser.User.Lastname,
					authenticatedUser.User.Email)
			} else {
				t.Logf("Aucun utilisateur authentifié")
			}

			// Créer le routeur de test
			router := testutils.CreateTestRouter()

			// Créer la requête HTTP
			req, err := http.NewRequest("GET", "/user/me", nil)
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter le token d'authentification si l'utilisateur est authentifié
			if authenticatedUser != nil && authenticatedUser.SessionToken != "" {
				req.Header.Set("Authorization", "Bearer "+authenticatedUser.SessionToken)
			}

			// Créer le recorder pour capturer la réponse
			w := httptest.NewRecorder()

			// Exécuter la requête
			router.ServeHTTP(w, req)

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, w.Code,
				"Code HTTP attendu %d, obtenu %d", testCase.ExpectedHttpCode, w.Code)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message,
					"Message attendu '%s', obtenu '%s'", testCase.ExpectedMessage, response.Message)
			}

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error,
					"Erreur attendue '%s', obtenue '%s'", testCase.ExpectedError, response.Error)
			}

			// Vérifications spécifiques pour les réponses réussies
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")
				require.NotNil(t, response.Data, "La réponse devrait contenir des données")

				// Vérifier que les données utilisateur sont retournées
				if data, ok := response.Data.(map[string]interface{}); ok {
					require.Contains(t, data, "user_id", "La réponse devrait contenir un user_id")
					require.Contains(t, data, "lastname", "La réponse devrait contenir un lastname")
					require.Contains(t, data, "firstname", "La réponse devrait contenir un firstname")
					require.Contains(t, data, "email", "La réponse devrait contenir un email")

					// Vérifier que les données correspondent à l'utilisateur authentifié
					if authenticatedUser != nil {
						require.Equal(t, float64(authenticatedUser.User.UserID), data["user_id"],
							"L'ID utilisateur devrait correspondre")
						require.Equal(t, authenticatedUser.User.Lastname, data["lastname"],
							"Le nom devrait correspondre")
						require.Equal(t, authenticatedUser.User.Firstname, data["firstname"],
							"Le prénom devrait correspondre")
						require.Equal(t, authenticatedUser.User.Email, data["email"],
							"L'email devrait correspondre")
					}
				}
			}

			// On purge les données après avoir traité le cas.
			testCase.CleanupData()
		})
	}
}

// TestUpdateUserMe teste la route PUT /user/me (mise à jour des données de l'utilisateur connecté)
func TestUpdateUserMe(t *testing.T) {
	// TestCases contient les cas qui seront testés
	var TestCases = []struct {
		CaseName         string
		SetupData        func() *testutils.AuthenticatedUser
		RequestData      func() common.UpdateUserRequest
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		CleanupData      func()
	}{
		{
			CaseName: "Mise à jour du nom avec utilisateur normal authentifié",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				nouveauNom := common.UpdateUserRequest{}
				lastname := "nouveauNom"
				nouveauNom.Lastname = &lastname

				return nouveauNom
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour du prénom avec utilisateur normal authentifié",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				newFirstname := "NouveauPrénom"
				return common.UpdateUserRequest{
					Firstname: &newFirstname,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour de l'email avec utilisateur normal authentifié",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				// Utiliser un email unique pour éviter les conflits
				uniqueUser, _ := testutils.GenerateAuthenticatedUser(false, false)
				newEmail := "update.email." + uniqueUser.User.Email
				return common.UpdateUserRequest{
					Email: &newEmail,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour du mot de passe avec utilisateur normal authentifié",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				newPassword := "NouveauMotDePasse123!"
				return common.UpdateUserRequest{
					Password: &newPassword,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour complète avec utilisateur normal authentifié",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				// Utiliser des emails uniques pour éviter les conflits
				uniqueUser, _ := testutils.GenerateAuthenticatedUser(false, false)
				return common.UpdateUserRequest{
					Lastname:  &uniqueUser.User.Lastname,
					Firstname: &uniqueUser.User.Firstname,
					Email:     &uniqueUser.User.Email,
					Password:  &uniqueUser.Password,
				}
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour avec email déjà utilisé par un autre utilisateur",
			SetupData: func() *testutils.AuthenticatedUser {
				// Créer d'abord un utilisateur avec un email fixe
				existingUser, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}

				// Créer ensuite l'utilisateur principal qui va essayer d'utiliser cet email
				mainUser, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}

				// Stocker l'email de l'utilisateur existant pour l'utiliser dans RequestData
				mainUser.User.Email = existingUser.User.Email
				return mainUser
			},
			RequestData: func() common.UpdateUserRequest {
				// Utiliser l'email de l'utilisateur existant créé dans SetupData
				conflictEmail := "conflict.email@test.example.com"
				return common.UpdateUserRequest{
					Email: &conflictEmail,
				}
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserAlreadyExists,
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour avec email invalide",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				invalidEmail := "email-invalide"
				return common.UpdateUserRequest{
					Email: &invalidEmail,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrInvalidEmailFormat,
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour avec mot de passe trop court",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				shortPassword := "123"
				return common.UpdateUserRequest{
					Password: &shortPassword,
				}
			},
			ExpectedHttpCode: http.StatusBadRequest,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrPasswordTooShort,
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Tentative de mise à jour sans authentification",
			SetupData: func() *testutils.AuthenticatedUser {
				return nil // Pas d'utilisateur authentifié
			},
			RequestData: func() common.UpdateUserRequest {
				newLastname := "Test"
				return common.UpdateUserRequest{
					Lastname: &newLastname,
				}
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    common.ErrUserNotAuthenticated,
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
		{
			CaseName: "Mise à jour avec données vides (requête valide)",
			SetupData: func() *testutils.AuthenticatedUser {
				user, err := testutils.GenerateAuthenticatedUser(true, true)
				if err != nil {
					return nil
				}
				return user
			},
			RequestData: func() common.UpdateUserRequest {
				return common.UpdateUserRequest{} // Aucune donnée à mettre à jour
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  common.MsgSuccessUserUpdate,
			ExpectedError:    "",
			CleanupData:      func() { /* Nettoyage automatique */ },
		},
	}

	// On boucle sur les cas de test contenu dans TestCases
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			// On isole le cas avant de le traiter.
			// On prépare les données utiles au traitement de ce cas.
			authenticatedUser := testCase.SetupData()

			// Vérifier si l'utilisateur a été créé correctement pour les cas qui en ont besoin
			if testCase.CaseName != "Tentative de mise à jour sans authentification" && authenticatedUser == nil {
				t.Fatal("Impossible de créer l'utilisateur de test")
			}

			// Obtenir les données de requête
			requestData := testCase.RequestData()

			if authenticatedUser != nil {
				t.Logf("Utilisateur de test: %s %s (%s)",
					authenticatedUser.User.Firstname,
					authenticatedUser.User.Lastname,
					authenticatedUser.User.Email)
				t.Logf("Données de mise à jour: Lastname=%v, Firstname=%v, Email=%v",
					getStringValue(requestData.Lastname),
					getStringValue(requestData.Firstname),
					getStringValue(requestData.Email))
			} else {
				t.Logf("Aucun utilisateur authentifié")
			}

			// Créer le routeur de test
			router := testutils.CreateTestRouter()

			// Préparer la requête JSON
			jsonData, err := json.Marshal(requestData)
			require.NoError(t, err, "Erreur lors de la sérialisation JSON")

			// Créer la requête HTTP
			req, err := http.NewRequest("PUT", "/user/me", bytes.NewBuffer(jsonData))
			require.NoError(t, err, "Erreur lors de la création de la requête")

			// Ajouter les headers nécessaires
			req.Header.Set("Content-Type", "application/json")

			// Ajouter le token d'authentification si l'utilisateur est authentifié
			if authenticatedUser != nil && authenticatedUser.SessionToken != "" {
				req.Header.Set("Authorization", "Bearer "+authenticatedUser.SessionToken)
			}

			// Créer le recorder pour capturer la réponse
			w := httptest.NewRecorder()

			// Exécuter la requête
			router.ServeHTTP(w, req)

			// Vérifier le code de statut HTTP
			require.Equal(t, testCase.ExpectedHttpCode, w.Code,
				"Code HTTP attendu %d, obtenu %d", testCase.ExpectedHttpCode, w.Code)

			// Parser la réponse JSON
			var response common.JSONResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Erreur lors du parsing de la réponse JSON")

			// Vérifier le message de succès si attendu
			if testCase.ExpectedMessage != "" {
				require.Equal(t, testCase.ExpectedMessage, response.Message,
					"Message attendu '%s', obtenu '%s'", testCase.ExpectedMessage, response.Message)
			}

			// Vérifier le message d'erreur si attendu
			if testCase.ExpectedError != "" {
				require.Equal(t, testCase.ExpectedError, response.Error,
					"Erreur attendue '%s', obtenue '%s'", testCase.ExpectedError, response.Error)
			}

			// Vérifications spécifiques pour les réponses réussies
			if testCase.ExpectedHttpCode == http.StatusOK {
				require.True(t, response.Success, "La réponse devrait indiquer un succès")

				// Optionnel : vérifier que les données ont été mises à jour en base
				// (cela nécessiterait une requête supplémentaire pour vérifier)
			}

			// On purge les données après avoir traité le cas.
			testCase.CleanupData()
		})
	}
}
