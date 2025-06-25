package calendar_event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-averroes/internal/calendar"
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/internal/session"
	"go-averroes/internal/user"
	"go-averroes/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	router := testutils.SetupTestRouter()

	// Configuration des routes pour les tests événements calendrier avec la nouvelle architecture
	// Routes publiques
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.POST("/auth/login", func(c *gin.Context) { session.Session.Login(c) })

	// Routes protégées par authentification
	router.POST("/calendar", middleware.AuthMiddleware(), func(c *gin.Context) { calendar.Calendar.Add(c) })
	router.GET("/calendar/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { calendar.Calendar.Get(c) })

	// Routes événements avec authentification et accès au calendrier
	router.GET("/calendar-event/:calendar_id/:event_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), middleware.EventExistsMiddleware("event_id"), func(c *gin.Context) { CalendarEvent.Get(c) })
	router.GET("/calendar-event/:calendar_id/month/:year/:month", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.ListByMonth(c) })
	router.GET("/calendar-event/:calendar_id/week/:year/:week", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.ListByWeek(c) })
	router.GET("/calendar-event/:calendar_id/day/:year/:month/:day", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.ListByDay(c) })
	router.GET("/calendar-event/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.List(c) })
	router.POST("/calendar-event/:calendar_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), func(c *gin.Context) { CalendarEvent.Add(c) })
	router.PUT("/calendar-event/:calendar_id/:event_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), middleware.EventExistsMiddleware("event_id"), func(c *gin.Context) { CalendarEvent.Update(c) })
	router.DELETE("/calendar-event/:calendar_id/:event_id", middleware.AuthMiddleware(), middleware.CalendarExistsMiddleware("calendar_id"), middleware.UserCanAccessCalendarMiddleware(), middleware.EventExistsMiddleware("event_id"), func(c *gin.Context) { CalendarEvent.Delete(c) })

	return router
}

// --- Helpers mutualisés pour la refonte table-driven ---
func createUser(router http.Handler, email, password, firstname, lastname string) (int, *http.Response) {
	payload := map[string]string{
		"email":     email,
		"password":  password,
		"firstname": firstname,
		"lastname":  lastname,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	var jsonResp struct {
		Success bool `json:"success"`
		Data    struct {
			UserID int `json:"user_id"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	return jsonResp.Data.UserID, resp
}

func loginAndGetToken(router http.Handler, email, password string) (string, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	var jsonResp struct {
		Success bool `json:"success"`
		Data    struct {
			SessionToken string `json:"session_token"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	if !jsonResp.Success || jsonResp.Data.SessionToken == "" {
		return "", fmt.Errorf("login failed: %s", jsonResp.Error)
	}
	return jsonResp.Data.SessionToken, nil
}

func createCalendar(router http.Handler, token, title, description string) (int, *http.Response) {
	payload := map[string]string{
		"title":       title,
		"description": description,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/calendar", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	var jsonResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			CalendarID int `json:"calendar_id"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	return jsonResp.Data.CalendarID, resp
}

func createEvent(router http.Handler, token string, calendarID int, title string, description string, start time.Time, duration int) (int, *http.Response) {
	payload := map[string]interface{}{
		"title":       title,
		"description": description,
		"start":       start.Format(time.RFC3339),
		"duration":    duration,
		"calendar_id": calendarID,
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("/calendar-event/%d", calendarID)
	req := httptest.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	var jsonResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			EventID int `json:"event_id"`
		} `json:"data"`
		Error string `json:"error"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &jsonResp)
	return jsonResp.Data.EventID, resp
}

func TestCalendarEvent_TableDriven(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()

	var TestCases = []struct {
		CaseName        string
		Method          string
		Setup           func() (token string, calendarID int, eventID int)
		URL             func(calendarID, eventID int) string
		Payload         func(calendarID int) string
		ExpectedStatus  int
		ExpectedSuccess bool
		ExpectedMsg     string
		ExpectedError   string
	}{
		// Création
		{
			CaseName: "Succès - Création événement",
			Method:   "POST",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				email := fmt.Sprintf("eventuser+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email, password, "Jean", "Event")
				token, _ := loginAndGetToken(router, email, password)
				calendarID, _ := createCalendar(router, token, "Calendrier Test", "Description")
				return token, calendarID, 0
			},
			URL: func(calendarID, _ int) string {
				return fmt.Sprintf("/calendar-event/%d", calendarID)
			},
			Payload: func(calendarID int) string {
				return fmt.Sprintf(`{"title":"Nouvel événement","description":"Desc","start":"%s","duration":60,"calendar_id":%d}`,
					time.Now().Add(24*time.Hour).Format(time.RFC3339), calendarID)
			},
			ExpectedStatus:  201,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessCreateEvent,
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Non authentifié (création)",
			Method:   "POST",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				email := fmt.Sprintf("eventuser+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email, password, "Jean", "Event")
				token, _ := loginAndGetToken(router, email, password)
				calendarID, _ := createCalendar(router, token, "Calendrier Test", "Description")
				return "", calendarID, 0
			},
			URL: func(calendarID, _ int) string {
				return fmt.Sprintf("/calendar-event/%d", calendarID)
			},
			Payload: func(calendarID int) string {
				return fmt.Sprintf(`{"title":"Nouvel événement","description":"Desc","start":"%s","duration":60,"calendar_id":%d}`,
					time.Now().Add(24*time.Hour).Format(time.RFC3339), calendarID)
			},
			ExpectedStatus:  401,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrUserNotAuthenticated,
		},
		{
			CaseName: "Erreur - Données invalides (titre manquant)",
			Method:   "POST",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				email := fmt.Sprintf("eventuser+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email, password, "Jean", "Event")
				token, _ := loginAndGetToken(router, email, password)
				calendarID, _ := createCalendar(router, token, "Calendrier Test", "Description")
				return token, calendarID, 0
			},
			URL: func(calendarID, _ int) string {
				return fmt.Sprintf("/calendar-event/%d", calendarID)
			},
			Payload: func(calendarID int) string {
				return fmt.Sprintf(`{"description":"Desc","start":"%s","duration":60,"calendar_id":%d}`,
					time.Now().Add(24*time.Hour).Format(time.RFC3339), calendarID)
			},
			ExpectedStatus:  400,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrInvalidData,
		},
		// Lecture
		{
			CaseName: "Succès - Get event",
			Method:   "GET",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				email := fmt.Sprintf("eventuser+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email, password, "Jean", "Event")
				token, _ := loginAndGetToken(router, email, password)
				calendarID, _ := createCalendar(router, token, "Calendrier Test", "Description")
				start := time.Now().Add(24 * time.Hour)
				eventID, _ := createEvent(router, token, calendarID, "Événement Test", "Description", start, 60)
				return token, calendarID, eventID
			},
			URL: func(calendarID, eventID int) string {
				return fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
			},
			Payload:         func(_ int) string { return "" },
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     "",
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Accès interdit (lecture)",
			Method:   "GET",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				// user1 crée event, user2 tente d'y accéder
				email1 := fmt.Sprintf("eventuser1+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email1, password, "Jean", "Event")
				token1, _ := loginAndGetToken(router, email1, password)
				calendarID, _ := createCalendar(router, token1, "Calendrier Test", "Description")
				start := time.Now().Add(24 * time.Hour)
				eventID, _ := createEvent(router, token1, calendarID, "Événement Test", "Description", start, 60)
				// user2
				email2 := fmt.Sprintf("eventuser2+%d@test.com", time.Now().UnixNano())
				_, _ = createUser(router, email2, password, "Autre", "User")
				token2, _ := loginAndGetToken(router, email2, password)
				return token2, calendarID, eventID
			},
			URL: func(calendarID, eventID int) string {
				return fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
			},
			Payload:         func(_ int) string { return "" },
			ExpectedStatus:  403,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrNoAccessToCalendar,
		},
		{
			CaseName: "Erreur - Event inexistant (lecture)",
			Method:   "GET",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				email := fmt.Sprintf("eventuser+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email, password, "Jean", "Event")
				token, _ := loginAndGetToken(router, email, password)
				calendarID, _ := createCalendar(router, token, "Calendrier Test", "Description")
				return token, calendarID, 99999
			},
			URL: func(calendarID, eventID int) string {
				return fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
			},
			Payload:         func(_ int) string { return "" },
			ExpectedStatus:  404,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrEventNotFound,
		},
		// Mise à jour
		{
			CaseName: "Succès - Update event",
			Method:   "PUT",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				email := fmt.Sprintf("eventuser+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email, password, "Jean", "Event")
				token, _ := loginAndGetToken(router, email, password)
				calendarID, _ := createCalendar(router, token, "Calendrier Test", "Description")
				start := time.Now().Add(24 * time.Hour)
				eventID, _ := createEvent(router, token, calendarID, "Événement Test", "Description", start, 60)
				return token, calendarID, eventID
			},
			URL: func(calendarID, eventID int) string {
				return fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
			},
			Payload: func(calendarID int) string {
				return fmt.Sprintf(`{"title":"Événement Modifié","duration":120,"calendar_id":%d}`, calendarID)
			},
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessUpdateEvent,
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Accès interdit (update)",
			Method:   "PUT",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				// user1 crée event, user2 tente d'update
				email1 := fmt.Sprintf("eventuser1+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email1, password, "Jean", "Event")
				token1, _ := loginAndGetToken(router, email1, password)
				calendarID, _ := createCalendar(router, token1, "Calendrier Test", "Description")
				start := time.Now().Add(24 * time.Hour)
				eventID, _ := createEvent(router, token1, calendarID, "Événement Test", "Description", start, 60)
				// user2
				email2 := fmt.Sprintf("eventuser2+%d@test.com", time.Now().UnixNano())
				_, _ = createUser(router, email2, password, "Autre", "User")
				token2, _ := loginAndGetToken(router, email2, password)
				return token2, calendarID, eventID
			},
			URL: func(calendarID, eventID int) string {
				return fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
			},
			Payload: func(calendarID int) string {
				return fmt.Sprintf(`{"title":"Hack","calendar_id":%d}`, calendarID)
			},
			ExpectedStatus:  403,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrNoAccessToCalendar,
		},
		// Suppression
		{
			CaseName: "Succès - Delete event",
			Method:   "DELETE",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				email := fmt.Sprintf("eventuser+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email, password, "Jean", "Event")
				token, _ := loginAndGetToken(router, email, password)
				calendarID, _ := createCalendar(router, token, "Calendrier Test", "Description")
				start := time.Now().Add(24 * time.Hour)
				eventID, _ := createEvent(router, token, calendarID, "Événement Test", "Description", start, 60)
				return token, calendarID, eventID
			},
			URL: func(calendarID, eventID int) string {
				return fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
			},
			Payload:         func(_ int) string { return "" },
			ExpectedStatus:  200,
			ExpectedSuccess: true,
			ExpectedMsg:     common.MsgSuccessDeleteEvent,
			ExpectedError:   "",
		},
		{
			CaseName: "Erreur - Accès interdit (delete)",
			Method:   "DELETE",
			Setup: func() (string, int, int) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				// user1 crée event, user2 tente de delete
				email1 := fmt.Sprintf("eventuser1+%d@test.com", time.Now().UnixNano())
				password := "motdepasse123"
				_, _ = createUser(router, email1, password, "Jean", "Event")
				token1, _ := loginAndGetToken(router, email1, password)
				calendarID, _ := createCalendar(router, token1, "Calendrier Test", "Description")
				start := time.Now().Add(24 * time.Hour)
				eventID, _ := createEvent(router, token1, calendarID, "Événement Test", "Description", start, 60)
				// user2
				email2 := fmt.Sprintf("eventuser2+%d@test.com", time.Now().UnixNano())
				_, _ = createUser(router, email2, password, "Autre", "User")
				token2, _ := loginAndGetToken(router, email2, password)
				return token2, calendarID, eventID
			},
			URL: func(calendarID, eventID int) string {
				return fmt.Sprintf("/calendar-event/%d/%d", calendarID, eventID)
			},
			Payload:         func(_ int) string { return "" },
			ExpectedStatus:  403,
			ExpectedSuccess: false,
			ExpectedMsg:     "",
			ExpectedError:   common.ErrNoAccessToCalendar,
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.CaseName, func(t *testing.T) {
			testutils.ResetTestDB()
			testutils.SetupTestDB()
			router := setupTestRouter()
			token, calendarID, eventID := testCase.Setup()
			var req *http.Request
			url := testCase.URL(calendarID, eventID)
			payload := testCase.Payload(calendarID)
			if testCase.Method == "GET" || testCase.Method == "DELETE" {
				req = httptest.NewRequest(testCase.Method, url, nil)
			} else {
				req = httptest.NewRequest(testCase.Method, url, strings.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
			}
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()
			var jsonResp struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			body, _ := io.ReadAll(resp.Body)
			_ = json.Unmarshal(body, &jsonResp)
			require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
			require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
			if testCase.ExpectedMsg != "" {
				require.Contains(t, jsonResp.Message, testCase.ExpectedMsg)
			}
			if testCase.ExpectedError != "" {
				require.Contains(t, jsonResp.Error, testCase.ExpectedError)
			}
		})
	}
}

func TestParseDateFilter_Day(t *testing.T) {
	start, end, err := parseDateFilter("day", "2024-01-15")
	if err != nil {
		t.Errorf("parseDateFilter(day) erreur inattendue: %v", err)
	}
	if !start.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Début incorrect: %v", start)
	}
	if !end.Equal(time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Fin incorrecte: %v", end)
	}
}

func TestParseDateFilter_Week(t *testing.T) {
	start, end, err := parseDateFilter("week", "2024-W01")
	if err != nil {
		t.Errorf("parseDateFilter(week) erreur inattendue: %v", err)
	}
	if !start.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Début incorrect: %v", start)
	}
	if !end.Equal(time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Fin incorrecte: %v", end)
	}
}

func TestParseDateFilter_Month(t *testing.T) {
	start, end, err := parseDateFilter("month", "2024-01")
	if err != nil {
		t.Errorf("parseDateFilter(month) erreur inattendue: %v", err)
	}
	if !start.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Début incorrect: %v", start)
	}
	if !end.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Fin incorrecte: %v", end)
	}
}

func TestParseDateFilter_Invalid(t *testing.T) {
	_, _, err := parseDateFilter("day", "invalid-date")
	if err == nil {
		t.Error("Attendu une erreur pour une date invalide")
	}
	_, _, err = parseDateFilter("week", "2024-WXX")
	if err == nil {
		t.Error("Attendu une erreur pour une semaine invalide")
	}
	_, _, err = parseDateFilter("month", "2024-XX")
	if err == nil {
		t.Error("Attendu une erreur pour un mois invalide")
	}
	_, _, err = parseDateFilter("unknown", "2024-01-01")
	if err == nil {
		t.Error("Attendu une erreur pour un type de filtre inconnu")
	}
}

func TestCalendarEvent_Listings(t *testing.T) {
	testutils.ResetTestDB()
	router := setupTestRouter()

	types := []struct {
		Name      string
		RouteGen  func(calendarID int, ref time.Time) string
		RefFormat func(ref time.Time) (year, month, day, week int)
	}{
		{
			Name: "month",
			RouteGen: func(calendarID int, ref time.Time) string {
				return fmt.Sprintf("/calendar-event/%d/month/%d/%d", calendarID, ref.Year(), int(ref.Month()))
			},
			RefFormat: func(ref time.Time) (int, int, int, int) {
				_, week := ref.ISOWeek()
				return ref.Year(), int(ref.Month()), ref.Day(), week
			},
		},
		{
			Name: "week",
			RouteGen: func(calendarID int, ref time.Time) string {
				_, week := ref.ISOWeek()
				return fmt.Sprintf("/calendar-event/%d/week/%d/%d", calendarID, ref.Year(), week)
			},
			RefFormat: func(ref time.Time) (int, int, int, int) {
				_, week := ref.ISOWeek()
				return ref.Year(), int(ref.Month()), ref.Day(), week
			},
		},
		{
			Name: "day",
			RouteGen: func(calendarID int, ref time.Time) string {
				return fmt.Sprintf("/calendar-event/%d/day/%d/%d/%d", calendarID, ref.Year(), int(ref.Month()), ref.Day())
			},
			RefFormat: func(ref time.Time) (int, int, int, int) {
				_, week := ref.ISOWeek()
				return ref.Year(), int(ref.Month()), ref.Day(), week
			},
		},
	}

	for _, typ := range types {
		t.Run("ListBy"+typ.Name, func(t *testing.T) {
			cases := []struct {
				CaseName        string
				Setup           func() (token string, calendarID int, ref time.Time, otherToken string)
				URL             func(calendarID int, ref time.Time) string
				Token           func(token, otherToken string) string
				ExpectedStatus  int
				ExpectedSuccess bool
				ExpectedMsg     string
				ExpectedError   string
				ExpectedCount   int
			}{
				{
					CaseName: "Succès - événements présents",
					Setup: func() (string, int, time.Time, string) {
						testutils.ResetTestDB()
						testutils.SetupTestDB()
						email := fmt.Sprintf("eventlist+%d@test.com", time.Now().UnixNano())
						password := "motdepasse123"
						_, _ = createUser(router, email, password, "Jean", "Event")
						token, _ := loginAndGetToken(router, email, password)
						calendarID, _ := createCalendar(router, token, "Calendrier Test", "Desc")
						// Utiliser une date UTC fixe pour éviter les problèmes de fuseau horaire
						ref := time.Date(2024, 6, 28, 10, 0, 0, 0, time.UTC)
						// Crée 2 événements dans la période
						_, _ = createEvent(router, token, calendarID, "Event1", "Desc1", ref, 60)
						_, _ = createEvent(router, token, calendarID, "Event2", "Desc2", ref.Add(2*time.Hour), 30)
						return token, calendarID, ref, ""
					},
					URL:             typ.RouteGen,
					Token:           func(token, _ string) string { return token },
					ExpectedStatus:  200,
					ExpectedSuccess: true,
					ExpectedMsg:     common.MsgSuccessListEvents,
					ExpectedError:   "",
					ExpectedCount:   2,
				},
				{
					CaseName: "Succès - aucun événement",
					Setup: func() (string, int, time.Time, string) {
						testutils.ResetTestDB()
						testutils.SetupTestDB()
						email := fmt.Sprintf("eventlist+%d@test.com", time.Now().UnixNano())
						password := "motdepasse123"
						_, _ = createUser(router, email, password, "Jean", "Event")
						token, _ := loginAndGetToken(router, email, password)
						calendarID, _ := createCalendar(router, token, "Calendrier Test", "Desc")
						ref := time.Now().Add(48 * time.Hour)
						return token, calendarID, ref, ""
					},
					URL:             typ.RouteGen,
					Token:           func(token, _ string) string { return token },
					ExpectedStatus:  200,
					ExpectedSuccess: true,
					ExpectedMsg:     common.MsgSuccessListEvents,
					ExpectedError:   "",
					ExpectedCount:   0,
				},
				{
					CaseName: "Erreur - Non authentifié",
					Setup: func() (string, int, time.Time, string) {
						testutils.ResetTestDB()
						testutils.SetupTestDB()
						email := fmt.Sprintf("eventlist+%d@test.com", time.Now().UnixNano())
						password := "motdepasse123"
						_, _ = createUser(router, email, password, "Jean", "Event")
						token, _ := loginAndGetToken(router, email, password)
						calendarID, _ := createCalendar(router, token, "Calendrier Test", "Desc")
						ref := time.Now().Add(48 * time.Hour)
						return "", calendarID, ref, ""
					},
					URL:             typ.RouteGen,
					Token:           func(_, _ string) string { return "" },
					ExpectedStatus:  401,
					ExpectedSuccess: false,
					ExpectedMsg:     "",
					ExpectedError:   common.ErrUserNotAuthenticated,
					ExpectedCount:   0,
				},
				{
					CaseName: "Erreur - Accès interdit",
					Setup: func() (string, int, time.Time, string) {
						testutils.ResetTestDB()
						testutils.SetupTestDB()
						email1 := fmt.Sprintf("eventlist1+%d@test.com", time.Now().UnixNano())
						password := "motdepasse123"
						_, _ = createUser(router, email1, password, "Jean", "Event")
						token1, _ := loginAndGetToken(router, email1, password)
						calendarID, _ := createCalendar(router, token1, "Calendrier Test", "Desc")
						ref := time.Now().Add(48 * time.Hour)
						_, _ = createEvent(router, token1, calendarID, "Event1", "Desc1", ref, 60)
						// user2
						email2 := fmt.Sprintf("eventlist2+%d@test.com", time.Now().UnixNano())
						_, _ = createUser(router, email2, password, "Autre", "User")
						otherToken, _ := loginAndGetToken(router, email2, password)
						return otherToken, calendarID, ref, token1
					},
					URL:             typ.RouteGen,
					Token:           func(token, _ string) string { return token },
					ExpectedStatus:  403,
					ExpectedSuccess: false,
					ExpectedMsg:     "",
					ExpectedError:   common.ErrNoAccessToCalendar,
					ExpectedCount:   0,
				},
				{
					CaseName: "Erreur - calendrier inexistant",
					Setup: func() (string, int, time.Time, string) {
						testutils.ResetTestDB()
						testutils.SetupTestDB()
						email := fmt.Sprintf("eventlist+%d@test.com", time.Now().UnixNano())
						password := "motdepasse123"
						_, _ = createUser(router, email, password, "Jean", "Event")
						token, _ := loginAndGetToken(router, email, password)
						ref := time.Now().Add(48 * time.Hour)
						return token, 99999, ref, ""
					},
					URL:             typ.RouteGen,
					Token:           func(token, _ string) string { return token },
					ExpectedStatus:  404,
					ExpectedSuccess: false,
					ExpectedMsg:     "",
					ExpectedError:   common.ErrCalendarNotFound,
					ExpectedCount:   0,
				},
				// Paramètres invalides (exemple pour ListByMonth)
			}

			for _, testCase := range cases {
				t.Run(testCase.CaseName, func(t *testing.T) {
					testutils.ResetTestDB()
					testutils.SetupTestDB()
					router := setupTestRouter()
					token, calendarID, ref, otherToken := testCase.Setup()
					url := testCase.URL(calendarID, ref)
					req := httptest.NewRequest("GET", url, nil)
					finalToken := testCase.Token(token, otherToken)
					if finalToken == "" {
						finalToken = token
					}
					if finalToken != "" {
						req.Header.Set("Authorization", "Bearer "+finalToken)
					}
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					resp := w.Result()
					defer resp.Body.Close()
					var jsonResp struct {
						Success bool                     `json:"success"`
						Message string                   `json:"message"`
						Error   string                   `json:"error"`
						Data    []map[string]interface{} `json:"data"`
					}
					body, _ := io.ReadAll(resp.Body)
					_ = json.Unmarshal(body, &jsonResp)
					require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
					require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
					if testCase.ExpectedMsg != "" {
						require.Contains(t, jsonResp.Message, testCase.ExpectedMsg)
					}
					if testCase.ExpectedError != "" {
						require.Contains(t, jsonResp.Error, testCase.ExpectedError)
					}
					if testCase.ExpectedStatus == 200 {
						require.Len(t, jsonResp.Data, testCase.ExpectedCount)
					}
				})
			}
		})
	}

	// Test de la route List (query params)
	t.Run("List (query params)", func(t *testing.T) {
		cases := []struct {
			CaseName        string
			Setup           func() (token string, calendarID int, ref time.Time)
			Query           func(ref time.Time) string
			Token           string
			ExpectedStatus  int
			ExpectedSuccess bool
			ExpectedMsg     string
			ExpectedError   string
			ExpectedCount   int
		}{
			{
				CaseName: "Succès - événements présents (month)",
				Setup: func() (string, int, time.Time) {
					testutils.ResetTestDB()
					testutils.SetupTestDB()
					email := fmt.Sprintf("eventlistq+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_, _ = createUser(router, email, password, "Jean", "Event")
					token, _ := loginAndGetToken(router, email, password)
					calendarID, _ := createCalendar(router, token, "Calendrier Test", "Desc")
					ref := time.Now().Add(48 * time.Hour)
					_, _ = createEvent(router, token, calendarID, "Event1", "Desc1", ref, 60)
					_, _ = createEvent(router, token, calendarID, "Event2", "Desc2", ref.Add(2*time.Hour), 30)
					return token, calendarID, ref
				},
				Query: func(ref time.Time) string {
					return fmt.Sprintf("filter_type=month&date=%d-%02d", ref.Year(), int(ref.Month()))
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessListEvents,
				ExpectedError:   "",
				ExpectedCount:   2,
			},
			{
				CaseName: "Succès - aucun événement (week)",
				Setup: func() (string, int, time.Time) {
					testutils.ResetTestDB()
					testutils.SetupTestDB()
					email := fmt.Sprintf("eventlistq+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_, _ = createUser(router, email, password, "Jean", "Event")
					token, _ := loginAndGetToken(router, email, password)
					calendarID, _ := createCalendar(router, token, "Calendrier Test", "Desc")
					ref := time.Now().Add(48 * time.Hour)
					return token, calendarID, ref
				},
				Query: func(ref time.Time) string {
					_, week := ref.ISOWeek()
					return fmt.Sprintf("filter_type=week&date=%d-W%02d", ref.Year(), week)
				},
				ExpectedStatus:  200,
				ExpectedSuccess: true,
				ExpectedMsg:     common.MsgSuccessListEvents,
				ExpectedError:   "",
				ExpectedCount:   0,
			},
			{
				CaseName: "Erreur - paramètres manquants",
				Setup: func() (string, int, time.Time) {
					testutils.ResetTestDB()
					testutils.SetupTestDB()
					email := fmt.Sprintf("eventlistq+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_, _ = createUser(router, email, password, "Jean", "Event")
					token, _ := loginAndGetToken(router, email, password)
					calendarID, _ := createCalendar(router, token, "Calendrier Test", "Desc")
					ref := time.Now().Add(48 * time.Hour)
					return token, calendarID, ref
				},
				Query: func(ref time.Time) string {
					return ""
				},
				ExpectedStatus:  400,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrMissingFilterParams,
				ExpectedCount:   0,
			},
			{
				CaseName: "Erreur - type de filtre invalide",
				Setup: func() (string, int, time.Time) {
					testutils.ResetTestDB()
					testutils.SetupTestDB()
					email := fmt.Sprintf("eventlistq+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_, _ = createUser(router, email, password, "Jean", "Event")
					token, _ := loginAndGetToken(router, email, password)
					calendarID, _ := createCalendar(router, token, "Calendrier Test", "Desc")
					ref := time.Now().Add(48 * time.Hour)
					return token, calendarID, ref
				},
				Query: func(ref time.Time) string {
					return "filter_type=invalid&date=2024-01-01"
				},
				ExpectedStatus:  400,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrInvalidFilterType,
				ExpectedCount:   0,
			},
			{
				CaseName: "Erreur - calendrier inexistant",
				Setup: func() (string, int, time.Time) {
					testutils.ResetTestDB()
					testutils.SetupTestDB()
					email := fmt.Sprintf("eventlistq+%d@test.com", time.Now().UnixNano())
					password := "motdepasse123"
					_, _ = createUser(router, email, password, "Jean", "Event")
					token, _ := loginAndGetToken(router, email, password)
					ref := time.Now().Add(48 * time.Hour)
					return token, 99999, ref
				},
				Query: func(ref time.Time) string {
					return fmt.Sprintf("filter_type=month&date=%d-%02d", ref.Year(), int(ref.Month()))
				},
				ExpectedStatus:  404,
				ExpectedSuccess: false,
				ExpectedMsg:     "",
				ExpectedError:   common.ErrCalendarNotFound,
				ExpectedCount:   0,
			},
		}

		for _, testCase := range cases {
			t.Run(testCase.CaseName, func(t *testing.T) {
				testutils.ResetTestDB()
				testutils.SetupTestDB()
				router := setupTestRouter()
				token, calendarID, ref := testCase.Setup()
				url := fmt.Sprintf("/calendar-event/%d?%s", calendarID, testCase.Query(ref))
				req := httptest.NewRequest("GET", url, nil)
				finalToken := testCase.Token
				if finalToken == "" {
					finalToken = token
				}
				if finalToken != "" {
					req.Header.Set("Authorization", "Bearer "+finalToken)
				}
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				resp := w.Result()
				defer resp.Body.Close()
				var jsonResp struct {
					Success bool                     `json:"success"`
					Message string                   `json:"message"`
					Error   string                   `json:"error"`
					Data    []map[string]interface{} `json:"data"`
				}
				body, _ := io.ReadAll(resp.Body)
				_ = json.Unmarshal(body, &jsonResp)
				require.Equal(t, testCase.ExpectedStatus, resp.StatusCode)
				require.Equal(t, testCase.ExpectedSuccess, jsonResp.Success)
				if testCase.ExpectedMsg != "" {
					require.Contains(t, jsonResp.Message, testCase.ExpectedMsg)
				}
				if testCase.ExpectedError != "" {
					require.Contains(t, jsonResp.Error, testCase.ExpectedError)
				}
				if testCase.ExpectedStatus == 200 {
					require.Len(t, jsonResp.Data, testCase.ExpectedCount)
				}
			})
		}
	})
}

func TestMain(m *testing.M) {
	testutils.SetupTestDB()
	code := m.Run()
	common.DB.Close()
	os.Exit(code)
}
