package testutils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-averroes/internal/calendar"
	"go-averroes/internal/calendar_event"
	"go-averroes/internal/common"
	"go-averroes/internal/middleware"
	"go-averroes/internal/role"
	"go-averroes/internal/session"
	"go-averroes/internal/user"
	"go-averroes/internal/user_calendar"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // Driver MySQL
	"golang.org/x/crypto/bcrypt"
)

func CreateTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Health check endpoint (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "GoLendar API",
		})
	})

	// ===== ROUTES D'AUTHENTIFICATION (publiques) =====
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", func(c *gin.Context) { session.Session.Login(c) })
		authGroup.POST("/refresh", func(c *gin.Context) { session.Session.RefreshToken(c) })
	}

	// ===== ROUTES D'AUTHENTIFICATION (protégées) =====
	authProtectedGroup := router.Group("/auth")
	authProtectedGroup.Use(middleware.AuthMiddleware())
	{
		authProtectedGroup.POST("/logout", func(c *gin.Context) { session.Session.Logout(c) })
		authProtectedGroup.GET("/me", func(c *gin.Context) { user.User.GetAuthMe(c) })
		authProtectedGroup.GET("/sessions", func(c *gin.Context) { session.Session.GetUserSessions(c) })
		authProtectedGroup.DELETE("/sessions/:session_id", func(c *gin.Context) { session.Session.DeleteSession(c) })
	}

	// ===== ROUTES DE GESTION DES UTILISATEURS =====
	userGroup := router.Group("/user")
	{
		// Création d'utilisateur (public - inscription)
		userGroup.POST("", func(c *gin.Context) { user.User.Add(c) })

		// Routes protégées par authentification
		userProtectedGroup := userGroup.Group("")
		userProtectedGroup.Use(middleware.AuthMiddleware())
		{
			// L'utilisateur peut accéder à ses propres données
			userProtectedGroup.GET("/me", func(c *gin.Context) { user.User.Get(c) })
			userProtectedGroup.PUT("/me", func(c *gin.Context) { user.User.Update(c) })
			userProtectedGroup.DELETE("/me", func(c *gin.Context) { user.User.Delete(c) })
		}

		// Routes admin pour gérer tous les utilisateurs
		userAdminGroup := userGroup.Group("")
		userAdminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			userAdminGroup.GET("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.Get(c) })
			userAdminGroup.PUT("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.Update(c) })
			userAdminGroup.DELETE("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.Delete(c) })
			userAdminGroup.GET("/:user_id/with-roles", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.GetUserWithRoles(c) })
		}
	}

	// ===== ROUTES DE GESTION DES RÔLES (admin uniquement) =====
	roleGroup := router.Group("/roles")
	roleGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		roleGroup.GET("", func(c *gin.Context) { role.Role.ListRoles(c) })
		roleGroup.GET("/:id", func(c *gin.Context) { role.Role.GetRole(c) })
		roleGroup.POST("", func(c *gin.Context) { role.Role.CreateRole(c) })
		roleGroup.PUT("/:id", func(c *gin.Context) { role.Role.UpdateRole(c) })
		roleGroup.DELETE("/:id", func(c *gin.Context) { role.Role.DeleteRole(c) })
		roleGroup.POST("/assign", func(c *gin.Context) { role.Role.AssignRole(c) })
		roleGroup.POST("/revoke", func(c *gin.Context) { role.Role.RevokeRole(c) })
		roleGroup.GET("/user/:user_id", func(c *gin.Context) { role.Role.GetUserRoles(c) })
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

	// ===== ROUTES DE GESTION DES CALENDRERS =====
	calendarGroup := router.Group("/calendar")
	calendarGroup.Use(middleware.AuthMiddleware())
	{
		// L'utilisateur peut créer des calendriers
		calendarGroup.POST("", func(c *gin.Context) { calendar.Calendar.Add(c) })

		// L'utilisateur peut accéder aux calendriers auxquels il a accès
		calendarGroup.GET("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar.Calendar.Get(c) },
		)
		calendarGroup.PUT("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar.Calendar.Update(c) },
		)
		calendarGroup.DELETE("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar.Calendar.Delete(c) },
		)
	}

	// ===== ROUTES DE GESTION DES ÉVÉNEMENTS =====
	calendarEventGroup := router.Group("/calendar-event")
	calendarEventGroup.Use(middleware.AuthMiddleware())
	{
		// Toutes les routes d'événements nécessitent l'accès au calendrier
		calendarEventGroup.GET("/:calendar_id/:event_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			middleware.EventExistsMiddleware("event_id"),
			func(c *gin.Context) { calendar_event.CalendarEvent.Get(c) },
		)
		calendarEventGroup.GET("/:calendar_id/month/:year/:month",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.ListByMonth(c) },
		)
		calendarEventGroup.GET("/:calendar_id/week/:year/:week",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.ListByWeek(c) },
		)
		calendarEventGroup.GET("/:calendar_id/day/:year/:month/:day",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.ListByDay(c) },
		)
		calendarEventGroup.POST("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.Add(c) },
		)
		calendarEventGroup.PUT("/:calendar_id/:event_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			middleware.EventExistsMiddleware("event_id"),
			func(c *gin.Context) { calendar_event.CalendarEvent.Update(c) },
		)
		calendarEventGroup.DELETE("/:calendar_id/:event_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			middleware.EventExistsMiddleware("event_id"),
			func(c *gin.Context) { calendar_event.CalendarEvent.Delete(c) },
		)
	}

	return router
}

// getEnv retourne la valeur d'une variable d'environnement ou une valeur par défaut si elle n'est pas définie.
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// InitTestDB initialise la base de données pour les tests
// Utilise les variables d'environnement avec des valeurs par défaut pour les tests
func InitTestDB() error {
	// Configuration de test avec variables d'environnement
	testConfig := common.DBConfig{
		User:     getEnv("TEST_DB_USER", "root"),
		Password: getEnv("TEST_DB_PASSWORD", "password"),
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     3306,                               // Port par défaut MySQL
		Name:     getEnv("TEST_DB_NAME", "calendar"), // Base de données de test séparée
	}

	return common.InitDB(testConfig)
}

// InitTestDBWithConfig initialise la base de données avec une configuration personnalisée
func InitTestDBWithConfig(config common.DBConfig) error {
	return common.InitDB(config)
}

// CleanupTestDB ferme la connexion à la base de données de test
func CleanupTestDB() error {
	if common.DB != nil {
		return common.DB.Close()
	}
	return nil
}

// SetupTestEnvironment configure l'environnement de test complet
func SetupTestEnvironment() error {
	// Initialiser la base de données de test
	if err := InitTestDB(); err != nil {
		return fmt.Errorf("erreur lors de l'initialisation de la base de données de test: %v", err)
	}

	// Ici on pourrait ajouter d'autres initialisations (logger, etc.)
	return nil
}

// TeardownTestEnvironment nettoie l'environnement de test
func TeardownTestEnvironment() error {
	// Fermer la connexion à la base de données
	if err := CleanupTestDB(); err != nil {
		return fmt.Errorf("erreur lors de la fermeture de la base de données de test: %v", err)
	}

	return nil
}

// GenerateUniqueEmail génère un email unique pour les tests
// Utilise un timestamp nanoseconde pour garantir l'unicité
func GenerateUniqueEmail(baseName string) string {
	return fmt.Sprintf("%s.%d@test.example.com", baseName, time.Now().UnixNano())
}

// Itoa convertit un int en string
func Itoa(i int) string {
	return strconv.Itoa(i)
}

// Purge toutes les données liées aux utilisateurs de test (user, user_roles, user_session, user_password, user_calendar)
func PurgeAllTestUsers() {
	if common.DB == nil {
		return
	}
	common.DB.Exec("SET FOREIGN_KEY_CHECKS=0;")
	common.DB.Exec("TRUNCATE TABLE user_calendar")
	common.DB.Exec("TRUNCATE TABLE user_roles")
	common.DB.Exec("TRUNCATE TABLE user_session")
	common.DB.Exec("TRUNCATE TABLE user_password")
	common.DB.Exec("TRUNCATE TABLE user")
	common.DB.Exec("SET FOREIGN_KEY_CHECKS=1;")
}

// AuthenticatedUser représente un utilisateur authentifié avec ses informations de session
type AuthenticatedUser struct {
	User         common.User
	Password     string
	SessionToken string
	RefreshToken string
	ExpiresAt    time.Time
	Roles        []common.Role
}

// GenerateAuthenticatedAdmin génère un admin avec option d'authentification
// Si authenticated = true, crée une session active d'une durée de 1 jour
// Si authenticated = false, crée seulement l'utilisateur admin sans session
func GenerateAuthenticatedAdmin(authenticated bool) (*AuthenticatedUser, error) {
	// Générer un email unique
	email := GenerateUniqueEmail("admin")
	password := "AdminPassword123!"

	// Créer l'utilisateur admin
	admin, err := createUserWithPassword("Admin", "User", email, password)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de l'admin: %v", err)
	}

	// Créer le rôle admin s'il n'existe pas
	adminRoleID, err := ensureAdminRoleExists()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du rôle admin: %v", err)
	}

	// Assigner le rôle admin à l'utilisateur
	err = assignRoleToUser(admin.UserID, adminRoleID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'attribution du rôle admin: %v", err)
	}

	// Récupérer les rôles de l'utilisateur
	roles, err := session.GetUserRoles(admin.UserID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des rôles: %v", err)
	}

	// Initialiser les valeurs par défaut pour un utilisateur non authentifié
	sessionToken := ""
	refreshToken := ""
	expiresAt := time.Time{}

	// Si authentifié, créer une session active d'une durée de 1 jour
	if authenticated {
		sessionToken, refreshToken, expiresAt, err = CreateUserSession(admin.UserID, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la création de la session: %v", err)
		}
	}

	return &AuthenticatedUser{
		User:         *admin,
		Password:     password,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Roles:        roles,
	}, nil
}

// GenerateAuthenticatedUser génère un utilisateur normal avec option d'authentification
// Si authenticated = true, crée une session active d'une durée de 1 jour
// Si authenticated = false, crée seulement l'utilisateur normal sans session
func GenerateAuthenticatedUser(authenticated bool) (*AuthenticatedUser, error) {
	// Générer un email unique
	email := GenerateUniqueEmail("user")
	password := "UserPassword123!"

	// Créer l'utilisateur normal
	user, err := createUserWithPassword("Normal", "User", email, password)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de l'utilisateur: %v", err)
	}

	// Récupérer les rôles de l'utilisateur (normalement vide pour un utilisateur normal)
	roles, err := session.GetUserRoles(user.UserID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des rôles: %v", err)
	}

	// Initialiser les valeurs par défaut pour un utilisateur non authentifié
	sessionToken := ""
	refreshToken := ""
	expiresAt := time.Time{}

	// Si authentifié, créer une session active d'une durée de 1 jour
	if authenticated {
		sessionToken, refreshToken, expiresAt, err = CreateUserSession(user.UserID, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la création de la session: %v", err)
		}
	}

	return &AuthenticatedUser{
		User:         *user,
		Password:     password,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Roles:        roles,
	}, nil
}

// Fonctions utilitaires privées

// createUserWithPassword crée un utilisateur avec un mot de passe hashé
func createUserWithPassword(lastname, firstname, email, password string) (*common.User, error) {
	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du hashage du mot de passe: %v", err)
	}

	tx, err := common.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erreur lors du démarrage de la transaction: %v", err)
	}
	defer tx.Rollback()

	// Créer l'utilisateur
	result, err := tx.Exec(`
		INSERT INTO user (lastname, firstname, email, created_at) 
		VALUES (?, ?, ?, NOW())
	`, lastname, firstname, email)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de l'utilisateur: %v", err)
	}

	userID, _ := result.LastInsertId()

	// Créer le mot de passe
	_, err = tx.Exec(`
		INSERT INTO user_password (user_id, password_hash, created_at) 
		VALUES (?, ?, NOW())
	`, userID, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du mot de passe: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erreur lors du commit de la transaction: %v", err)
	}

	// Récupérer l'utilisateur créé
	var user common.User
	err = common.DB.QueryRow(`
		SELECT user_id, lastname, firstname, email, created_at, updated_at, deleted_at
		FROM user WHERE user_id = ?
	`, userID).Scan(
		&user.UserID,
		&user.Lastname,
		&user.Firstname,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur: %v", err)
	}

	return &user, nil
}

// ensureAdminRoleExists s'assure que le rôle admin existe et retourne son ID
func ensureAdminRoleExists() (int, error) {
	// Vérifier si le rôle admin existe déjà
	var roleID int
	err := common.DB.QueryRow("SELECT role_id FROM roles WHERE name = 'admin' AND deleted_at IS NULL").Scan(&roleID)
	if err == nil {
		// Le rôle existe déjà
		return roleID, nil
	}

	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("erreur lors de la vérification du rôle admin: %v", err)
	}

	// Le rôle n'existe pas, le créer
	result, err := common.DB.Exec(`
		INSERT INTO roles (name, description, created_at) 
		VALUES ('admin', 'Administrateur avec tous les droits', NOW())
	`)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la création du rôle admin: %v", err)
	}

	newRoleID, _ := result.LastInsertId()
	return int(newRoleID), nil
}

// assignRoleToUser assigne un rôle à un utilisateur
func assignRoleToUser(userID, roleID int) error {
	// Vérifier si l'attribution existe déjà
	var existingID int
	err := common.DB.QueryRow("SELECT user_roles_id FROM user_roles WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL", userID, roleID).Scan(&existingID)
	if err == nil {
		// L'attribution existe déjà
		return nil
	}

	if err != sql.ErrNoRows {
		return fmt.Errorf("erreur lors de la vérification de l'attribution de rôle: %v", err)
	}

	// Créer l'attribution
	_, err = common.DB.Exec(`
		INSERT INTO user_roles (user_id, role_id, created_at) 
		VALUES (?, ?, NOW())
	`, userID, roleID)
	if err != nil {
		return fmt.Errorf("erreur lors de l'attribution du rôle: %v", err)
	}

	return nil
}

// CreateUserSession crée une session pour un utilisateur avec une durée spécifiée
func CreateUserSession(userID int, duration time.Duration) (string, string, time.Time, error) {
	// Générer les tokens
	sessionToken, err := generateToken()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("erreur lors de la génération du session token: %v", err)
	}

	refreshToken, err := generateToken()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("erreur lors de la génération du refresh token: %v", err)
	}

	// Définir l'expiration
	expiresAt := time.Now().Add(duration)

	// Créer la session en base
	_, err = common.DB.Exec(`
		INSERT INTO user_session (user_id, session_token, refresh_token, expires_at, device_info, ip_address, is_active, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, TRUE, NOW())
	`, userID, sessionToken, refreshToken, expiresAt, "Test Device", "127.0.0.1")
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("erreur lors de la création de la session: %v", err)
	}

	return sessionToken, refreshToken, expiresAt, nil
}

// generateToken génère un token aléatoire (copié depuis session.go)
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
