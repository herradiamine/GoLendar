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
		RequestData      func() common.CreateUserRequest
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		SetupData        func() string
		CleanupData      func()
	}{
		{
			CaseName: "Création d'utilisateur avec données valides",
			RequestData: func() common.CreateUserRequest {
				return common.CreateUserRequest{
					Lastname:  "Dupont",
					Firstname: "Jean",
					Email:     "jean.dupont@example.com",
					Password:  "MotDePasse123!",
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
			RequestData: func() common.CreateUserRequest {
				existingUser, _ := testutils.GenerateAuthenticatedUser(false, true)
				// L'email sera défini après le setup
				return common.CreateUserRequest{
					Lastname:  "Martin",
					Firstname: "Marie",
					Email:     existingUser.User.Email, // Sera défini dynamiquement
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
			RequestData: func() common.CreateUserRequest {
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
			RequestData: func() common.CreateUserRequest {
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
			RequestData: func() common.CreateUserRequest {
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
			RequestData: func() common.CreateUserRequest {
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
			RequestData: func() common.CreateUserRequest {
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
			RequestData: func() common.CreateUserRequest {
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
