package user_calendar_test

import (
	"encoding/json"
	"go-averroes/internal/middleware"
	"go-averroes/internal/session"
	"go-averroes/internal/user_calendar"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func createTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// ===== ROUTES D'AUTHENTIFICATION (publiques) =====
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", func(c *gin.Context) { session.Session.Login(c) })
		authGroup.POST("/refresh", func(c *gin.Context) { session.Session.RefreshToken(c) })
	}

	// ===== ROUTES DE GESTION DES LIAISONS USER-CALENDAR (admin uniquement) =====
	userCalendarGroup := router.Group("/user-calendar")
	userCalendarGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		userCalendarGroup.GET("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Get(c) },
		)
		userCalendarGroup.GET("/:user_id",
			middleware.UserExistsMiddleware("user_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.List(c) },
		)
		userCalendarGroup.POST("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Add(c) },
		)
		userCalendarGroup.PUT("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Update(c) },
		)
		userCalendarGroup.DELETE("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Delete(c) },
		)
	}

	return router
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
	var TestCases = []struct {
		CaseName         string
		SetupData        func() (adminToken string, userID int, calendarID int, cleanup func())
		ExpectedHttpCode int
		ExpectedMessage  string
		ExpectedError    string
	}{
		{
			CaseName: "Création de liaison user-calendar (succès)",
			SetupData: func() (string, int, int, func()) {
				// Créer un admin, un user, un calendar
				admin, adminToken, _ := testutils.CreateAdminUser(1001, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				user := testutils.CreateTestUser(1002, testutils.GenerateUniqueEmail("user"))
				calendarID := testutils.CreateTestCalendar()
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
			SetupData: func() (string, int, int, func()) {
				admin, adminToken, _ := testutils.CreateAdminUser(1003, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				user := testutils.CreateTestUser(1004, testutils.GenerateUniqueEmail("user"))
				calendarID := testutils.CreateTestCalendar()
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
			ExpectedError:    "Liaison déjà existante", // À adapter
		},
		{
			CaseName: "Utilisateur inexistant",
			SetupData: func() (string, int, int, func()) {
				admin, adminToken, _ := testutils.CreateAdminUser(1005, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				calendarID := testutils.CreateTestCalendar()
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
			SetupData: func() (string, int, int, func()) {
				admin, adminToken, _ := testutils.CreateAdminUser(1006, "Admin", "Test", testutils.GenerateUniqueEmail("admin"))
				user := testutils.CreateTestUser(1007, testutils.GenerateUniqueEmail("user"))
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
			SetupData: func() (string, int, int, func()) {
				user := testutils.CreateTestUser(1008, testutils.GenerateUniqueEmail("user"))
				calendarID := testutils.CreateTestCalendar()
				_, token, _ := testutils.CreateAuthenticatedUser(user.UserID, user.Lastname, user.Firstname, user.Email)
				cleanup := func() {
					_ = testutils.PurgeTestData(user.Email)
					_ = testutils.PurgeTestCalendar(calendarID)
				}
				return "Bearer " + token, user.UserID, calendarID, cleanup
			},
			ExpectedHttpCode: http.StatusForbidden,
			ExpectedMessage:  "",
			ExpectedError:    "Accès refusé", // À adapter
		},
	}

	router := createTestRouter()
	gin.SetMode(gin.TestMode)

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			token, userID, calendarID, cleanup := testCase.SetupData()
			defer cleanup()

			url := "/user-calendar/" + testutils.Itoa(userID) + "/" + testutils.Itoa(calendarID)
			req, err := http.NewRequest("POST", url, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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
