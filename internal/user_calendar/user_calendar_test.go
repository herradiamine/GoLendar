package user_calendar_test

import (
	"encoding/json"
	"go-averroes/testutils"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

// TestAddUserCalendar teste la création d'une liaison user-calendar
func TestAddUserCalendar(t *testing.T) {
	testutils.PurgeAllTestUsers()
	router := testutils.CreateTestRouter()
	gin.SetMode(gin.TestMode)

	var TestCases = []struct {
		CaseName         string
		SetupDataWithIDs func(adminID, userID int) (string, int, int, func())
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Création de liaison user-calendar (succès)",
			SetupDataWithIDs: func(adminID, userID int) (string, int, int, func()) {
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user")+"-"+testutils.Itoa(int(time.Now().UnixNano())), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusCreated,
			ExpectedMessage:  "Liaison utilisateur-calendrier créée avec succès", // À adapter
			ExpectedError:    "",
		},
		{
			CaseName: "Conflit : liaison déjà existante",
			SetupDataWithIDs: func(adminID, userID int) (string, int, int, func()) {
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user")+"-"+testutils.Itoa(int(time.Now().UnixNano())), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				_ = testutils.AddUserCalendarLink(user.UserID, calendarID)
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusConflict,
			ExpectedMessage:  "",
			ExpectedError:    "Liaison utilisateur-calendrier déjà existante", // À adapter
		},
		{
			CaseName: "Utilisateur inexistant",
			SetupDataWithIDs: func(adminID, userID int) (string, int, int, func()) {
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, 999999, calendarID, cleanup // user inexistant
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    "Utilisateur non trouvé", // À adapter
		},
		{
			CaseName: "Calendrier inexistant",
			SetupDataWithIDs: func(adminID, userID int) (string, int, int, func()) {
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user")+"-"+testutils.Itoa(int(time.Now().UnixNano())), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
				}
				return "Bearer " + adminToken, user.UserID, 999999, cleanup // calendar inexistant
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    "Calendrier non trouvé", // À adapter
		},
		{
			CaseName: "Accès non admin refusé",
			SetupDataWithIDs: func(adminID, userID int) (string, int, int, func()) {
				// Générer deux emails et deux userID uniques
				userID1 := int(time.Now().UnixNano() % 1000000000)
				userID2 := userID1 + rand.Intn(1000000) + 1
				email1 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID1)
				email2 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID2)
				user, err := testutils.CreateUserWithPassword("Test", "User", email1, "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				_, token, err := testutils.CreateAuthenticatedUser(userID2, user.Lastname, user.Firstname, email2)
				if err != nil {
					panic("Erreur création session: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(email1)
					_ = testutils.PurgeTestData(email2)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + token, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    "Permissions insuffisantes", // Message exact retourné par l'API
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			testutils.PurgeAllTestUsers() // Purge avant chaque test
			adminID := int(time.Now().UnixNano() % 1000000000)
			userID := int(time.Now().UnixNano() % 1000000000)
			token, userID, calendarID, cleanup := testCase.SetupDataWithIDs(adminID, userID)
			defer cleanup()

			url := "/user-calendar/" + testutils.Itoa(userID) + "/" + testutils.Itoa(calendarID)
			req, err := http.NewRequest("POST", url, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != testCase.ExpectedHttpCode {
				t.Logf("Body de la réponse (code %d): %s", w.Code, w.Body.String())
			}
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)

			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if testCase.ExpectedError != "" {
				require.Contains(t, response["error"], testCase.ExpectedError)
			}
			if testCase.ExpectedMessage != "" {
				require.Contains(t, response["message"], testCase.ExpectedMessage)
			}
		})
	}
}

func TestGetUserCalendar(t *testing.T) {
	testutils.PurgeAllTestUsers()
	gin.SetMode(gin.TestMode)
	router := testutils.CreateTestRouter()

	var TestCases = []struct {
		CaseName         string
		SetupData        func() (token string, userID int, calendarID int, cleanup func())
		ExpectedHttpCode int
		ExpectedError    string
	}{
		{
			CaseName: "Succès (admin)",
			SetupData: func() (string, int, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedError:    "",
		},
		{
			CaseName: "Liaison inexistante",
			SetupData: func() (string, int, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedError:    "Liaison utilisateur-calendrier non trouvée",
		},
		{
			CaseName: "Utilisateur non admin (accès refusé)",
			SetupData: func() (string, int, int, func()) {
				userID1 := int(time.Now().UnixNano() % 1000000000)
				userID2 := userID1 + rand.Intn(1000000) + 1
				email1 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID1)
				email2 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID2)
				user, err := testutils.CreateUserWithPassword("Test", "User", email1, "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				_, token, err := testutils.CreateAuthenticatedUser(userID2, user.Lastname, user.Firstname, email2)
				if err != nil {
					panic("Erreur création session: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(email1)
					_ = testutils.PurgeTestData(email2)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + token, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedError:    "Permissions insuffisantes",
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupData: func() (string, int, int, func()) {
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "", user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedError:    "Utilisateur non authentifié",
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userID, calendarID, cleanup := testCase.SetupData()
			defer cleanup()
			url := "/user-calendar/" + testutils.Itoa(userID) + "/" + testutils.Itoa(calendarID)
			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != testCase.ExpectedHttpCode {
				t.Logf("Body de la réponse (code %d): %s", w.Code, w.Body.String())
			}
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)
			if testCase.ExpectedError != "" {
				var response map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				require.Contains(t, response["error"], testCase.ExpectedError)
			}
		})
	}
}

func TestUpdateUserCalendar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var TestCases = []struct {
		CaseName         string
		SetupData        func() (token string, userID int, calendarID int, cleanup func())
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Succès (admin)",
			SetupData: func() (string, int, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Liaison utilisateur-calendrier mise à jour avec succès",
			ExpectedError:    "",
		},
		{
			CaseName: "Liaison inexistante",
			SetupData: func() (string, int, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    "Liaison utilisateur-calendrier non trouvée",
		},
		{
			CaseName: "Utilisateur non admin (accès refusé)",
			SetupData: func() (string, int, int, func()) {
				userID1 := int(time.Now().UnixNano() % 1000000000)
				userID2 := userID1 + rand.Intn(1000000) + 1
				email1 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID1)
				email2 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID2)
				user, err := testutils.CreateUserWithPassword("Test", "User", email1, "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				_, token, err := testutils.CreateAuthenticatedUser(userID2, user.Lastname, user.Firstname, email2)
				if err != nil {
					panic("Erreur création session: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(email1)
					_ = testutils.PurgeTestData(email2)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + token, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    "Permissions insuffisantes",
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupData: func() (string, int, int, func()) {
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "", user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "Utilisateur non authentifié",
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userID, calendarID, cleanup := testCase.SetupData()
			defer cleanup()
			url := "/user-calendar/" + testutils.Itoa(userID) + "/" + testutils.Itoa(calendarID)
			req, err := http.NewRequest("PUT", url, nil)
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}
			w := httptest.NewRecorder()
			router := testutils.CreateTestRouter()
			router.ServeHTTP(w, req)
			if w.Code != testCase.ExpectedHttpCode {
				t.Logf("Body de la réponse (code %d): %s", w.Code, w.Body.String())
			}
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)
			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)
			if testCase.ExpectedError != "" {
				require.Contains(t, response["error"], testCase.ExpectedError)
			}
			if testCase.ExpectedMessage != "" {
				require.Contains(t, response["message"], testCase.ExpectedMessage)
			}
		})
	}
}

func TestDeleteUserCalendar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var TestCases = []struct {
		CaseName         string
		SetupData        func() (token string, userID int, calendarID int, cleanup func())
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Succès (admin)",
			SetupData: func() (string, int, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Liaison utilisateur-calendrier supprimée avec succès",
			ExpectedError:    "",
		},
		{
			CaseName: "Liaison inexistante",
			SetupData: func() (string, int, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + adminToken, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusNotFound,
			ExpectedMessage:  "",
			ExpectedError:    "Liaison utilisateur-calendrier non trouvée",
		},
		{
			CaseName: "Utilisateur non admin (accès refusé)",
			SetupData: func() (string, int, int, func()) {
				userID1 := int(time.Now().UnixNano() % 1000000000)
				userID2 := userID1 + rand.Intn(1000000) + 1
				email1 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID1)
				email2 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID2)
				user, err := testutils.CreateUserWithPassword("Test", "User", email1, "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				_, token, err := testutils.CreateAuthenticatedUser(userID2, user.Lastname, user.Firstname, email2)
				if err != nil {
					panic("Erreur création session: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(email1)
					_ = testutils.PurgeTestData(email2)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + token, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    "Permissions insuffisantes",
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupData: func() (string, int, int, func()) {
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "", user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "Utilisateur non authentifié",
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userID, calendarID, cleanup := testCase.SetupData()
			defer cleanup()
			url := "/user-calendar/" + testutils.Itoa(userID) + "/" + testutils.Itoa(calendarID)
			req, err := http.NewRequest("DELETE", url, nil)
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}
			w := httptest.NewRecorder()
			router := testutils.CreateTestRouter()
			router.ServeHTTP(w, req)
			if w.Code != testCase.ExpectedHttpCode {
				t.Logf("Body de la réponse (code %d): %s", w.Code, w.Body.String())
			}
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)
			var response map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &response)
			if testCase.ExpectedError != "" {
				require.Contains(t, response["error"], testCase.ExpectedError)
			}
			if testCase.ExpectedMessage != "" {
				require.Contains(t, response["message"], testCase.ExpectedMessage)
			}
		})
	}
}

func TestListUserCalendars(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var TestCases = []struct {
		CaseName         string
		SetupData        func() (token string, userID int, cleanup func())
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		ExpectedCount    int // nombre de liaisons attendues
	}{
		{
			CaseName: "Succès (admin authentifié, plusieurs liaisons)",
			SetupData: func() (string, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				calendarID1, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				calendarID2, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(admin.UserID, calendarID1); err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(admin.UserID, calendarID2); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestCalendar(calendarID1)
					_ = testutils.PurgeTestCalendar(calendarID2)
				}
				return "Bearer " + adminToken, admin.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Liste des calendriers récupérée avec succès",
			ExpectedError:    "",
			ExpectedCount:    2,
		},
		{
			CaseName: "Utilisateur sans liaisons",
			SetupData: func() (string, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
				}
				return "Bearer " + adminToken, user.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Liste des calendriers récupérée avec succès",
			ExpectedError:    "",
			ExpectedCount:    0,
		},
		{
			CaseName: "Utilisateur non admin (accès refusé)",
			SetupData: func() (string, int, func()) {
				userID1 := int(time.Now().UnixNano() % 1000000000)
				userID2 := userID1 + rand.Intn(1000000) + 1
				email1 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID1)
				email2 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID2)
				user, err := testutils.CreateUserWithPassword("Test", "User", email1, "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				_, token, err := testutils.CreateAuthenticatedUser(userID2, user.Lastname, user.Firstname, email2)
				if err != nil {
					panic("Erreur création session: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(email1)
					_ = testutils.PurgeTestData(email2)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + token, user.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    "Permissions insuffisantes",
			ExpectedCount:    0,
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupData: func() (string, int, func()) {
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "", user.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "Utilisateur non authentifié",
			ExpectedCount:    0,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userID, cleanup := testCase.SetupData()
			defer cleanup()
			url := "/user-calendar/" + testutils.Itoa(userID)
			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}
			w := httptest.NewRecorder()
			router := testutils.CreateTestRouter()
			router.ServeHTTP(w, req)
			if w.Code != testCase.ExpectedHttpCode {
				t.Logf("Body de la réponse (code %d): %s", w.Code, w.Body.String())
			}
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)
			var response map[string]json.RawMessage
			_ = json.Unmarshal(w.Body.Bytes(), &response)
			if testCase.ExpectedError != "" {
				var errMsg string
				_ = json.Unmarshal(response["error"], &errMsg)
				require.Contains(t, errMsg, testCase.ExpectedError)
			}
			if testCase.ExpectedMessage != "" {
				var msg string
				_ = json.Unmarshal(response["message"], &msg)
				require.Contains(t, msg, testCase.ExpectedMessage)
			}
			if testCase.ExpectedHttpCode == http.StatusOK && testCase.ExpectedError == "" {
				var data []interface{}
				_ = json.Unmarshal(response["data"], &data)
				require.Equal(t, testCase.ExpectedCount, len(data))
			}
		})
	}
}

func TestGetUserCalendarsByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var TestCases = []struct {
		CaseName         string
		SetupData        func() (token string, userID int, cleanup func())
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
		ExpectedCount    int
	}{
		{
			CaseName: "Succès (admin, plusieurs liaisons)",
			SetupData: func() (string, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				calendarID1, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				calendarID2, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(admin.UserID, calendarID1); err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(admin.UserID, calendarID2); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestCalendar(calendarID1)
					_ = testutils.PurgeTestCalendar(calendarID2)
				}
				return "Bearer " + adminToken, admin.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Liste des calendriers récupérée avec succès",
			ExpectedError:    "",
			ExpectedCount:    2,
		},
		{
			CaseName: "Utilisateur sans liaisons",
			SetupData: func() (string, int, func()) {
				adminID := int(time.Now().UnixNano() % 1000000000)
				admin, adminToken, err := testutils.CreateAdminUser(adminID, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				if err != nil {
					panic("Erreur création admin: " + err.Error())
				}
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(admin.Email)
					_ = testutils.PurgeTestData(user.Email)
				}
				return "Bearer " + adminToken, user.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusOK,
			ExpectedMessage:  "Liste des calendriers récupérée avec succès",
			ExpectedError:    "",
			ExpectedCount:    0,
		},
		{
			CaseName: "Utilisateur non admin (accès refusé)",
			SetupData: func() (string, int, func()) {
				userID1 := int(time.Now().UnixNano() % 1000000000)
				userID2 := userID1 + rand.Intn(1000000) + 1
				email1 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID1)
				email2 := testutils.GenerateUniqueEmail("user") + "-" + testutils.Itoa(userID2)
				user, err := testutils.CreateUserWithPassword("Test", "User", email1, "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				_, token, err := testutils.CreateAuthenticatedUser(userID2, user.Lastname, user.Firstname, email2)
				if err != nil {
					panic("Erreur création session: " + err.Error())
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(email1)
					_ = testutils.PurgeTestData(email2)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + token, user.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    "Permissions insuffisantes",
			ExpectedCount:    0,
		},
		{
			CaseName: "Utilisateur non authentifié",
			SetupData: func() (string, int, func()) {
				user, err := testutils.CreateUserWithPassword("Test", "User", testutils.GenerateUniqueEmail("user"), "password123")
				if err != nil {
					panic("Erreur création user: " + err.Error())
				}
				calendarID, err := testutils.CreateTestCalendar()
				if err != nil {
					panic(err)
				}
				if err := testutils.AddUserCalendarLink(user.UserID, calendarID); err != nil {
					panic(err)
				}
				cleanup := func() {
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "", user.UserID, cleanup
			},
			ExpectedHttpCode: http.StatusUnauthorized,
			ExpectedMessage:  "",
			ExpectedError:    "Utilisateur non authentifié",
			ExpectedCount:    0,
		},
	}
	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userID, cleanup := testCase.SetupData()
			defer cleanup()
			url := "/user-calendar/" + testutils.Itoa(userID)
			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)
			if token != "" {
				req.Header.Set("Authorization", token)
			}
			w := httptest.NewRecorder()
			router := testutils.CreateTestRouter()
			router.ServeHTTP(w, req)
			if w.Code != testCase.ExpectedHttpCode {
				t.Logf("Body de la réponse (code %d): %s", w.Code, w.Body.String())
			}
			require.Equal(t, testCase.ExpectedHttpCode, w.Code)
			var response map[string]json.RawMessage
			_ = json.Unmarshal(w.Body.Bytes(), &response)
			if testCase.ExpectedError != "" {
				var errMsg string
				_ = json.Unmarshal(response["error"], &errMsg)
				require.Contains(t, errMsg, testCase.ExpectedError)
			}
			if testCase.ExpectedMessage != "" {
				var msg string
				_ = json.Unmarshal(response["message"], &msg)
				require.Contains(t, msg, testCase.ExpectedMessage)
			}
			if testCase.ExpectedHttpCode == http.StatusOK && testCase.ExpectedError == "" {
				var data []interface{}
				_ = json.Unmarshal(response["data"], &data)
				require.Equal(t, testCase.ExpectedCount, len(data))
			}
		})
	}
}
