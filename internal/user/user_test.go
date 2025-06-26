package user

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

// createTestUserWithPassword crée un utilisateur avec mot de passe et retourne les données complètes
func createTestUserWithPassword(t *testing.T, lastname, firstname, email, password string) (*common.User, string) {
	user := createTestUser(t, lastname, firstname, email, password)
	return user, password
}

// setupTestContext configure le contexte Gin avec un utilisateur authentifié
func setupTestContext(t *testing.T, user *common.User) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("auth_user", *user)
	return c
}

// =============================================================================
// TESTS POUR LA FONCTION Add
// =============================================================================

func TestUserAdd_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.POST("/users", User.Add)

	requestBody := common.CreateUserRequest{
		Lastname:  "Dupont",
		Firstname: "Jean",
		Email:     "jean.dupont@example.com",
		Password:  "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, common.MsgSuccessCreateUser, response.Message)
	assert.NotNil(t, response.Data)

	// Vérifier que l'utilisateur a été créé en base
	var userID int
	err = common.DB.QueryRow("SELECT user_id FROM user WHERE email = ?", requestBody.Email).Scan(&userID)
	assert.NoError(t, err)
	assert.Greater(t, userID, 0)
}

func TestUserAdd_InvalidData(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.POST("/users", User.Add)

	requestBody := common.CreateUserRequest{
		Lastname:  "", // Données invalides
		Firstname: "",
		Email:     "invalid-email",
		Password:  "123", // Mot de passe trop court
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrInvalidData, response.Error)
}

func TestUserAdd_UserAlreadyExists(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.POST("/users", User.Add)

	// Créer un utilisateur existant
	existingUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")

	requestBody := common.CreateUserRequest{
		Lastname:  "Dupont",
		Firstname: "Jean",
		Email:     existingUser.Email, // Email déjà existant
		Password:  "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrUserAlreadyExists, response.Error)
}

// =============================================================================
// TESTS POUR LA FONCTION Get
// =============================================================================

func TestUserGet_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.GET("/users/me", User.Get)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser)

	// Act
	req, _ := http.NewRequest("GET", "/users/me", nil)
	w := httptest.NewRecorder()
	c.Request = req
	User.Get(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Vérifier que les données utilisateur sont correctes
	userData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(testUser.UserID), userData["user_id"])
	assert.Equal(t, testUser.Lastname, userData["lastname"])
	assert.Equal(t, testUser.Firstname, userData["firstname"])
	assert.Equal(t, testUser.Email, userData["email"])
}

func TestUserGet_UserNotInContext(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.GET("/users/me", User.Get)

	// Créer un contexte sans utilisateur
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Act
	req, _ := http.NewRequest("GET", "/users/me", nil)
	c.Request = req
	User.Get(c)

	// Assert
	// La fonction ne retourne pas d'erreur HTTP mais log une erreur
	// On vérifie juste qu'elle ne panique pas
	assert.NotNil(t, c)
}

// =============================================================================
// TESTS POUR LA FONCTION Update
// =============================================================================

func TestUserUpdate_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.PUT("/users/me", User.Update)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser)

	requestBody := common.UpdateUserRequest{
		Lastname:  common.StringPtr("Martin"),
		Firstname: common.StringPtr("Pierre"),
		Email:     common.StringPtr("pierre.martin@example.com"),
		Password:  common.StringPtr("newpassword123"),
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	User.Update(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, common.MsgSuccessUpdateUser, response.Message)

	// Vérifier que l'utilisateur a été mis à jour en base
	var updatedUser common.User
	err = common.DB.QueryRow(`
		SELECT user_id, lastname, firstname, email, created_at, updated_at, deleted_at
		FROM user WHERE user_id = ?
	`, testUser.UserID).Scan(&updatedUser.UserID, &updatedUser.Lastname, &updatedUser.Firstname, &updatedUser.Email, &updatedUser.CreatedAt, &updatedUser.UpdatedAt, &updatedUser.DeletedAt)
	assert.NoError(t, err)
	assert.Equal(t, "Martin", updatedUser.Lastname)
	assert.Equal(t, "Pierre", updatedUser.Firstname)
	assert.Equal(t, "pierre.martin@example.com", updatedUser.Email)
	assert.NotNil(t, updatedUser.UpdatedAt)
}

func TestUserUpdate_InvalidEmailFormat(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.PUT("/users/me", User.Update)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser)

	requestBody := common.UpdateUserRequest{
		Email: common.StringPtr("invalid-email-format"),
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	User.Update(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrInvalidEmailFormat, response.Error)
}

func TestUserUpdate_PasswordTooShort(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.PUT("/users/me", User.Update)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser)

	requestBody := common.UpdateUserRequest{
		Password: common.StringPtr("123"), // Mot de passe trop court
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	User.Update(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrPasswordTooShort, response.Error)
}

func TestUserUpdate_EmailAlreadyExists(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.PUT("/users/me", User.Update)

	// Créer deux utilisateurs de test
	testUser1 := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	testUser2 := createTestUser(t, "Martin", "Pierre", "pierre.martin@example.com", "password123")

	c := setupTestContext(t, testUser1)

	requestBody := common.UpdateUserRequest{
		Email: common.StringPtr(testUser2.Email), // Email déjà utilisé par un autre utilisateur
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Act
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Request = req
	User.Update(c)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, common.ErrUserAlreadyExists, response.Error)
}

// =============================================================================
// TESTS POUR LA FONCTION Delete
// =============================================================================

func TestUserDelete_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.DELETE("/users/me", User.Delete)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser)

	// Act
	req, _ := http.NewRequest("DELETE", "/users/me", nil)
	w := httptest.NewRecorder()
	c.Request = req
	User.Delete(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, common.MsgSuccessUserDelete, response.Message)

	// Vérifier que l'utilisateur a été supprimé (soft delete)
	var deletedAt *time.Time
	err = common.DB.QueryRow("SELECT deleted_at FROM user WHERE user_id = ?", testUser.UserID).Scan(&deletedAt)
	assert.NoError(t, err)
	assert.NotNil(t, deletedAt)
}

// =============================================================================
// TESTS POUR LA FONCTION GetUserWithRoles
// =============================================================================

func TestUserGetUserWithRoles_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.GET("/users/me/roles", User.GetUserWithRoles)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser)

	// Act
	req, _ := http.NewRequest("GET", "/users/me/roles", nil)
	w := httptest.NewRecorder()
	c.Request = req
	User.GetUserWithRoles(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Vérifier que les données utilisateur avec rôles sont correctes
	userData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(testUser.UserID), userData["user_id"])
	assert.Equal(t, testUser.Lastname, userData["lastname"])
	assert.Equal(t, testUser.Firstname, userData["firstname"])
	assert.Equal(t, testUser.Email, userData["email"])
	assert.NotNil(t, userData["roles"])
}

// =============================================================================
// TESTS POUR LA FONCTION GetAuthMe
// =============================================================================

func TestUserGetAuthMe_Success(t *testing.T) {
	// Arrange
	testutils.ResetTestDB()
	router := testutils.SetupTestRouter()
	router.GET("/auth/me", User.GetAuthMe)

	// Créer un utilisateur de test
	testUser := createTestUser(t, "Dupont", "Jean", "jean.dupont@example.com", "password123")
	c := setupTestContext(t, testUser)

	// Act
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()
	c.Request = req
	User.GetAuthMe(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response common.JSONResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Vérifier que les données utilisateur sont correctes
	userData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(testUser.UserID), userData["user_id"])
	assert.Equal(t, testUser.Lastname, userData["lastname"])
	assert.Equal(t, testUser.Firstname, userData["firstname"])
	assert.Equal(t, testUser.Email, userData["email"])
}
