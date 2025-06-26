package calendar

import (
	"bytes"
	"encoding/json"
	"go-averroes/internal/common"
	"go-averroes/testutils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// =============================================================================
// FONCTIONS MUTUALISÉES POUR LES DONNÉES DE TEST
// =============================================================================

// createTestUser crée un utilisateur de test dans la base de données
func createTestUser(t *testing.T, lastname, firstname, email, password string) *common.User {
	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Insérer l'utilisateur
	result, err := common.DB.Exec(`
		INSERT INTO user (lastname, firstname, email, created_at) 
		VALUES (?, ?, ?, NOW())
	`, lastname, firstname, email)
	assert.NoError(t, err)

	userID, err := result.LastInsertId()
	assert.NoError(t, err)

	// Insérer le mot de passe
	_, err = common.DB.Exec(`
		INSERT INTO user_password (user_id, password_hash, created_at) 
		VALUES (?, ?, NOW())
	`, userID, string(hashedPassword))
	assert.NoError(t, err)

	// Récupérer l'utilisateur créé
	var user common.User
	err = common.DB.QueryRow(`
		SELECT user_id, lastname, firstname, email, created_at, updated_at, deleted_at
		FROM user WHERE user_id = ?
	`, userID).Scan(&user.UserID, &user.Lastname, &user.Firstname, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	assert.NoError(t, err)

	return &user
}

// createTestCalendar crée un calendrier de test dans la base de données
func createTestCalendar(t *testing.T, title string, description *string) *common.Calendar {
	// Insérer le calendrier
	result, err := common.DB.Exec(`
		INSERT INTO calendar (title, description, created_at) 
		VALUES (?, ?, NOW())
	`, title, description)
	assert.NoError(t, err)

	calendarID, err := result.LastInsertId()
	assert.NoError(t, err)

	// Récupérer le calendrier créé
	var calendar common.Calendar
	err = common.DB.QueryRow(`
		SELECT calendar_id, title, description, created_at, updated_at, deleted_at
		FROM calendar WHERE calendar_id = ?
	`, calendarID).Scan(&calendar.CalendarID, &calendar.Title, &calendar.Description, &calendar.CreatedAt, &calendar.UpdatedAt, &calendar.DeletedAt)
	assert.NoError(t, err)

	return &calendar
}

// createTestUserCalendar crée une liaison utilisateur-calendrier de test
func createTestUserCalendar(t *testing.T, userID int, calendarID int) *common.UserCalendar {
	// Insérer la liaison
	result, err := common.DB.Exec(`
		INSERT INTO user_calendar (user_id, calendar_id, created_at) 
		VALUES (?, ?, NOW())
	`, userID, calendarID)
	assert.NoError(t, err)

	userCalendarID, err := result.LastInsertId()
	assert.NoError(t, err)

	// Récupérer la liaison créée
	var userCalendar common.UserCalendar
	err = common.DB.QueryRow(`
		SELECT user_calendar_id, user_id, calendar_id, created_at, updated_at, deleted_at
		FROM user_calendar WHERE user_calendar_id = ?
	`, userCalendarID).Scan(&userCalendar.UserCalendarID, &userCalendar.UserID, &userCalendar.CalendarID, &userCalendar.CreatedAt, &userCalendar.UpdatedAt, &userCalendar.DeletedAt)
	assert.NoError(t, err)

	return &userCalendar
}

// setupTestContext configure le contexte Gin avec un utilisateur et un calendrier
func setupTestContext(t *testing.T, user *common.User, calendar *common.Calendar) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("auth_user", *user)
	if calendar != nil {
		c.Set("calendar", *calendar)
	}
	return c
}

// =============================================================================
// TESTS POUR LA FONCTION Add
// =============================================================================

func TestCalendarAdd_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.POST("/calendars", Calendar.Add)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser, nil)

	requestBody := common.CreateCalendarRequest{
		Title:       "Mon Calendrier Personnel",
		Description: common.StringPtr("Description de mon calendrier"),
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("POST", "/calendars", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Add(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, common.MsgSuccessCreateCalendar, response.Message)
	assert.NotNil(t, response.Data)

	// Vérifier que le calendrier a été créé en base
	var calendarID int
	err = common.DB.QueryRow("SELECT calendar_id FROM calendar WHERE title = ?", requestBody.Title).Scan(&calendarID)
	assert.NoError(t, err)
	assert.Greater(t, calendarID, 0)

	// Vérifier que la liaison utilisateur-calendrier a été créée
	var userCalendarID int
	err = common.DB.QueryRow("SELECT user_calendar_id FROM user_calendar WHERE user_id = ? AND calendar_id = ?", testUser.UserID, calendarID).Scan(&userCalendarID)
	assert.NoError(t, err)
	assert.Greater(t, userCalendarID, 0)
}

func TestCalendarAdd_InvalidData(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.POST("/calendars", Calendar.Add)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser, nil)

	requestBody := common.CreateCalendarRequest{
		Title: "", // Titre vide - données invalides
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("POST", "/calendars", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Add(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, common.ErrInvalidData)
}

func TestCalendarAdd_UserNotAuthenticated(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.POST("/calendars", Calendar.Add)

	// Créer un contexte sans utilisateur authentifié
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	requestBody := common.CreateCalendarRequest{
		Title:       "Mon Calendrier Personnel",
		Description: common.StringPtr("Description de mon calendrier"),
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("POST", "/calendars", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Add(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrUserNotAuthenticated, response.Error)
}

// =============================================================================
// TESTS POUR LA FONCTION Get
// =============================================================================

func TestCalendarGet_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.GET("/calendars/:id", Calendar.Get)

	// Créer un utilisateur et un calendrier de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	testCalendar := createTestCalendar(t, "Mon Calendrier", common.StringPtr("Description"))
	createTestUserCalendar(t, testUser.UserID, testCalendar.CalendarID)

	c := setupTestContext(t, testUser, testCalendar)

	// Act
	req, _ := http.NewRequest("GET", "/calendars/1", nil)
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Get(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, common.MsgSuccessGetCalendar, response.Message)
	assert.NotNil(t, response.Data)

	// Vérifier que les données du calendrier sont correctes
	calendarData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(testCalendar.CalendarID), calendarData["calendar_id"])
	assert.Equal(t, testCalendar.Title, calendarData["title"])
	assert.Equal(t, *testCalendar.Description, calendarData["description"])
}

func TestCalendarGet_UserNotAuthenticated(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.GET("/calendars/:id", Calendar.Get)

	// Créer un contexte sans utilisateur authentifié
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Act
	req, _ := http.NewRequest("GET", "/calendars/1", nil)
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Get(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrUserNotAuthenticated, response.Error)
}

func TestCalendarGet_CalendarNotInContext(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.GET("/calendars/:id", Calendar.Get)

	// Créer un utilisateur de test sans calendrier dans le contexte
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser, nil)

	// Act
	req, _ := http.NewRequest("GET", "/calendars/1", nil)
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Get(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrCalendarNotFound, response.Error)
}

// =============================================================================
// TESTS POUR LA FONCTION Update
// =============================================================================

func TestCalendarUpdate_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.PUT("/calendars/:id", Calendar.Update)

	// Créer un utilisateur et un calendrier de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	testCalendar := createTestCalendar(t, "Mon Calendrier", common.StringPtr("Description"))
	createTestUserCalendar(t, testUser.UserID, testCalendar.CalendarID)

	c := setupTestContext(t, testUser, testCalendar)

	requestBody := common.UpdateCalendarRequest{
		Title:       common.StringPtr("Mon Calendrier Modifié"),
		Description: common.StringPtr("Nouvelle description"),
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("PUT", "/calendars/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Update(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, common.MsgSuccessUpdateCalendar, response.Message)

	// Vérifier que le calendrier a été mis à jour en base
	var updatedCalendar common.Calendar
	err = common.DB.QueryRow(`
		SELECT calendar_id, title, description, created_at, updated_at, deleted_at
		FROM calendar WHERE calendar_id = ?
	`, testCalendar.CalendarID).Scan(&updatedCalendar.CalendarID, &updatedCalendar.Title, &updatedCalendar.Description, &updatedCalendar.CreatedAt, &updatedCalendar.UpdatedAt, &updatedCalendar.DeletedAt)
	assert.NoError(t, err)
	assert.Equal(t, "Mon Calendrier Modifié", updatedCalendar.Title)
	assert.Equal(t, "Nouvelle description", *updatedCalendar.Description)
	assert.NotNil(t, updatedCalendar.UpdatedAt)
}

func TestCalendarUpdate_InvalidData(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.PUT("/calendars/:id", Calendar.Update)

	// Créer un utilisateur et un calendrier de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	testCalendar := createTestCalendar(t, "Mon Calendrier", common.StringPtr("Description"))
	createTestUserCalendar(t, testUser.UserID, testCalendar.CalendarID)

	c := setupTestContext(t, testUser, testCalendar)

	requestBody := common.UpdateCalendarRequest{
		Title: common.StringPtr(""), // Titre vide - données invalides
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("PUT", "/calendars/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Update(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrInvalidData, response.Error)
}

func TestCalendarUpdate_UserNotAuthenticated(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.PUT("/calendars/:id", Calendar.Update)

	// Créer un contexte sans utilisateur authentifié
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	requestBody := common.UpdateCalendarRequest{
		Title:       common.StringPtr("Mon Calendrier Modifié"),
		Description: common.StringPtr("Nouvelle description"),
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("PUT", "/calendars/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Update(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrUserNotAuthenticated, response.Error)
}

// =============================================================================
// TESTS POUR LA FONCTION Delete
// =============================================================================

func TestCalendarDelete_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.DELETE("/calendars/:id", Calendar.Delete)

	// Créer un utilisateur et un calendrier de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	testCalendar := createTestCalendar(t, "Mon Calendrier", common.StringPtr("Description"))
	createTestUserCalendar(t, testUser.UserID, testCalendar.CalendarID)

	c := setupTestContext(t, testUser, testCalendar)

	// Act
	req, _ := http.NewRequest("DELETE", "/calendars/1", nil)
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Delete(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, common.MsgSuccessDeleteCalendar, response.Message)

	// Vérifier que le calendrier a été supprimé (soft delete)
	var deletedAt *time.Time
	err = common.DB.QueryRow("SELECT deleted_at FROM calendar WHERE calendar_id = ?", testCalendar.CalendarID).Scan(&deletedAt)
	assert.NoError(t, err)
	assert.NotNil(t, deletedAt)
}

func TestCalendarDelete_UserNotAuthenticated(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.DELETE("/calendars/:id", Calendar.Delete)

	// Créer un contexte sans utilisateur authentifié
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Act
	req, _ := http.NewRequest("DELETE", "/calendars/1", nil)
	w := httptest.NewRecorder()
	c.Request = req
	Calendar.Delete(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrUserNotAuthenticated, response.Error)
}
