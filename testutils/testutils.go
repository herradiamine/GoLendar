package testutils

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
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
	PurgeAllTestUsers()
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

	// S'assurer que le rôle standard 'user' existe
	_, _ = common.DB.Exec(`
		INSERT INTO roles (name, description, created_at)
		VALUES (?, ?, NOW())
		ON DUPLICATE KEY UPDATE updated_at = NOW()
	`, "user", "Utilisateur standard avec accès à ses propres calendriers et événements")

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

// GenerateUniqueEmail génère un email unique avec des lettres aléatoires
// La longueur totale de l'email ne dépasse pas 320 caractères
func GenerateUniqueEmail(baseName string) string {
	// Constantes pour la génération
	const (
		domain    = "@test.example.com"
		separator = "."
		maxLength = 320
		randomLen = 8 // Longueur de la partie aléatoire
	)

	// Calculer la longueur maximale disponible pour baseName
	maxBaseNameLength := maxLength - len(domain) - len(separator) - randomLen

	// Tronquer baseName si nécessaire
	if len(baseName) > maxBaseNameLength {
		baseName = baseName[:maxBaseNameLength]
	}

	// Générer des lettres aléatoires
	randomPart := generateRandomLetters(randomLen)

	// Construire l'email
	email := baseName + separator + randomPart + domain

	// Vérification de sécurité (ne devrait jamais dépasser 320)
	if len(email) > maxLength {
		// En cas de dépassement, tronquer davantage
		excess := len(email) - maxLength
		if len(baseName) > excess {
			baseName = baseName[:len(baseName)-excess]
		}
		email = baseName + separator + randomPart + domain
	}

	return email
}

// generateRandomLetters génère une chaîne de lettres aléatoires
func generateRandomLetters(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[mathrand.Intn(len(letters))]
	}
	return string(result)
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
	common.DB.Exec("TRUNCATE TABLE calendar_event")
	common.DB.Exec("TRUNCATE TABLE event")
	common.DB.Exec("TRUNCATE TABLE user_calendar")
	common.DB.Exec("TRUNCATE TABLE calendar")
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
	Calendar     *common.Calendar
	Event        *common.Event
}

// GenerateAuthenticatedAdmin génère un admin avec option d'authentification
// Si authenticated = true, crée une session active d'une durée de 1 jour
// Si authenticated = false, crée seulement l'utilisateur admin sans session
// Si saveToDB = true, enregistre l'utilisateur et la session en base de données
// Si saveToDB = false, crée seulement l'objet en mémoire sans persistance
// Si hasCalendar = true, crée un calendrier associé à l'utilisateur
// Si hasEvent = true, crée un événement dans le calendrier (nécessite hasCalendar = true)
func GenerateAuthenticatedAdmin(authenticated, saveToDB, hasCalendar, hasEvent bool) (*AuthenticatedUser, error) {
	// Générer un email unique
	email := GenerateUniqueEmail("admin")
	password := "AdminPassword123!"

	var admin *common.User
	var roles []common.Role
	var sessionToken, refreshToken string
	var expiresAt time.Time
	var calendar *common.Calendar
	var event *common.Event
	var err error

	if saveToDB {
		// Créer l'utilisateur admin en base
		admin, err = createUserWithPassword("Admin", "User", email, password)
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
		roles, err = session.GetUserRoles(admin.UserID)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la récupération des rôles: %v", err)
		}

		// Si authentifié, créer une session active d'une durée de 1 jour
		if authenticated {
			sessionToken, refreshToken, expiresAt, err = CreateUserSession(admin.UserID, 24*time.Hour)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la création de la session: %v", err)
			}
		}

		// Créer un calendrier si demandé
		if hasCalendar {
			calendarID, err := CreateCalendarForUser(admin.UserID, "Calendrier Admin", "Calendrier de test pour admin")
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la création du calendrier: %v", err)
			}

			// Récupérer les informations du calendrier créé
			calendar = &common.Calendar{}
			err = common.DB.QueryRow(`
				SELECT calendar_id, title, description, created_at, updated_at, deleted_at 
				FROM calendar 
				WHERE calendar_id = ?
			`, calendarID).Scan(&calendar.CalendarID, &calendar.Title, &calendar.Description, &calendar.CreatedAt, &calendar.UpdatedAt, &calendar.DeletedAt)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la récupération du calendrier: %v", err)
			}
		}

		// Créer un événement si demandé (nécessite un calendrier)
		if hasEvent && hasCalendar {
			eventID, err := createEventForUser(admin.UserID, "Événement Admin", "Événement de test pour admin")
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la création de l'événement: %v", err)
			}

			// Récupérer les informations de l'événement créé
			event = &common.Event{}
			err = common.DB.QueryRow(`
				SELECT event_id, title, description, start, duration, canceled, created_at, updated_at, deleted_at 
				FROM event 
				WHERE event_id = ?
			`, eventID).Scan(&event.EventID, &event.Title, &event.Description, &event.Start, &event.Duration, &event.Canceled, &event.CreatedAt, &event.UpdatedAt, &event.DeletedAt)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la récupération de l'événement: %v", err)
			}
		}
	} else {
		// Créer l'utilisateur admin en mémoire seulement
		admin = &common.User{
			UserID:    1, // ID fictif pour les tests
			Lastname:  "Admin",
			Firstname: "User",
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: nil,
			DeletedAt: nil,
		}

		// Créer le rôle admin en mémoire
		adminRole := common.Role{
			RoleID:      1,
			Name:        "admin",
			Description: common.StringPtr("Administrateur avec tous les droits"),
			CreatedAt:   time.Now(),
			UpdatedAt:   nil,
			DeletedAt:   nil,
		}
		roles = []common.Role{adminRole}

		// Si authentifié, générer des tokens en mémoire
		if authenticated {
			sessionToken, err = generateToken()
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la génération du session token: %v", err)
			}

			refreshToken, err = generateToken()
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la génération du refresh token: %v", err)
			}

			expiresAt = time.Now().Add(24 * time.Hour)
		}

		// Créer un calendrier en mémoire si demandé
		if hasCalendar {
			calendar = &common.Calendar{
				CalendarID:  1,
				Title:       "Calendrier Admin",
				Description: common.StringPtr("Calendrier de test pour admin"),
				CreatedAt:   time.Now(),
				UpdatedAt:   nil,
				DeletedAt:   nil,
			}
		}

		// Créer un événement en mémoire si demandé
		if hasEvent && hasCalendar {
			event = &common.Event{
				EventID:     1,
				Title:       "Événement Admin",
				Description: common.StringPtr("Événement de test pour admin"),
				Start:       time.Now().Add(1 * time.Hour),
				Duration:    60,
				Canceled:    false,
				CreatedAt:   time.Now(),
				UpdatedAt:   nil,
				DeletedAt:   nil,
			}
		}
	}

	return &AuthenticatedUser{
		User:         *admin,
		Password:     password,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Roles:        roles,
		Calendar:     calendar,
		Event:        event,
	}, nil
}

// GenerateAuthenticatedUser génère un utilisateur normal avec option d'authentification
// Si authenticated = true, crée une session active d'une durée de 1 jour
// Si authenticated = false, crée seulement l'utilisateur normal sans session
// Si saveToDB = true, enregistre l'utilisateur et la session en base de données
// Si saveToDB = false, crée seulement l'objet en mémoire sans persistance
// Si hasCalendar = true, crée un calendrier associé à l'utilisateur
// Si hasEvent = true, crée un événement dans le calendrier (nécessite hasCalendar = true)
func GenerateAuthenticatedUser(authenticated, saveToDB, hasCalendar, hasEvent bool) (*AuthenticatedUser, error) {
	// Générer un email unique
	email := GenerateUniqueEmail("user")
	password := "UserPassword123!"

	var user *common.User
	var roles []common.Role
	var sessionToken, refreshToken string
	var expiresAt time.Time
	var calendar *common.Calendar
	var event *common.Event
	var err error

	if saveToDB {
		// Créer l'utilisateur normal en base
		user, err = createUserWithPassword("Normal", "User", email, password)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la création de l'utilisateur: %v", err)
		}

		// Récupérer les rôles de l'utilisateur (normalement vide pour un utilisateur normal)
		roles, err = session.GetUserRoles(user.UserID)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la récupération des rôles: %v", err)
		}

		// Si authentifié, créer une session active d'une durée de 1 jour
		if authenticated {
			sessionToken, refreshToken, expiresAt, err = CreateUserSession(user.UserID, 24*time.Hour)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la création de la session: %v", err)
			}
		}

		// Créer un calendrier si demandé
		if hasCalendar {
			calendarID, err := CreateCalendarForUser(user.UserID, "Calendrier Utilisateur", "Calendrier de test pour utilisateur")
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la création du calendrier: %v", err)
			}

			// Récupérer les informations du calendrier créé
			calendar = &common.Calendar{}
			err = common.DB.QueryRow(`
				SELECT calendar_id, title, description, created_at, updated_at, deleted_at 
				FROM calendar 
				WHERE calendar_id = ?
			`, calendarID).Scan(&calendar.CalendarID, &calendar.Title, &calendar.Description, &calendar.CreatedAt, &calendar.UpdatedAt, &calendar.DeletedAt)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la récupération du calendrier: %v", err)
			}
		}

		// Créer un événement si demandé (nécessite un calendrier)
		if hasEvent && hasCalendar {
			eventID, err := createEventForUser(user.UserID, "Événement Utilisateur", "Événement de test pour utilisateur")
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la création de l'événement: %v", err)
			}

			// Récupérer les informations de l'événement créé
			event = &common.Event{}
			err = common.DB.QueryRow(`
				SELECT event_id, title, description, start, duration, canceled, created_at, updated_at, deleted_at 
				FROM event 
				WHERE event_id = ?
			`, eventID).Scan(&event.EventID, &event.Title, &event.Description, &event.Start, &event.Duration, &event.Canceled, &event.CreatedAt, &event.UpdatedAt, &event.DeletedAt)
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la récupération de l'événement: %v", err)
			}
		}
	} else {
		// Créer l'utilisateur normal en mémoire seulement
		user = &common.User{
			UserID:    2, // ID fictif pour les tests
			Lastname:  "Normal",
			Firstname: "User",
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: nil,
			DeletedAt: nil,
		}

		// Utilisateur normal sans rôles
		roles = []common.Role{}

		// Si authentifié, générer des tokens en mémoire
		if authenticated {
			sessionToken, err = generateToken()
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la génération du session token: %v", err)
			}

			refreshToken, err = generateToken()
			if err != nil {
				return nil, fmt.Errorf("erreur lors de la génération du refresh token: %v", err)
			}

			expiresAt = time.Now().Add(24 * time.Hour)
		}

		// Créer un calendrier en mémoire si demandé
		if hasCalendar {
			calendar = &common.Calendar{
				CalendarID:  2,
				Title:       "Calendrier Utilisateur",
				Description: common.StringPtr("Calendrier de test pour utilisateur"),
				CreatedAt:   time.Now(),
				UpdatedAt:   nil,
				DeletedAt:   nil,
			}
		}

		// Créer un événement en mémoire si demandé
		if hasEvent && hasCalendar {
			event = &common.Event{
				EventID:     2,
				Title:       "Événement Utilisateur",
				Description: common.StringPtr("Événement de test pour utilisateur"),
				Start:       time.Now().Add(1 * time.Hour),
				Duration:    60,
				Canceled:    false,
				CreatedAt:   time.Now(),
				UpdatedAt:   nil,
				DeletedAt:   nil,
			}
		}
	}

	return &AuthenticatedUser{
		User:         *user,
		Password:     password,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Roles:        roles,
		Calendar:     calendar,
		Event:        event,
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
		INSERT INTO user_session (user_id, session_token, refresh_token, expires_at, device_info, ip_address, location, is_active, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, TRUE, NOW())
	`, userID, sessionToken, refreshToken, expiresAt, "Test Device", "127.0.0.1", "Local")
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

// GetStringValue retourne la valeur d'un pointeur string ou "<nil>" si nil
func GetStringValue(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

// createCalendarForUser crée un calendrier et l'associe à un utilisateur
func CreateCalendarForUser(userID int, title, description string) (int, error) {
	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("erreur lors du démarrage de la transaction: %v", err)
	}
	defer tx.Rollback()

	// Créer le calendrier
	result, err := tx.Exec(`
		INSERT INTO calendar (title, description, created_at) 
		VALUES (?, ?, NOW())
	`, title, description)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la création du calendrier: %v", err)
	}

	calendarID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération de l'ID du calendrier: %v", err)
	}

	// Associer le calendrier à l'utilisateur
	_, err = tx.Exec(`
		INSERT INTO user_calendar (user_id, calendar_id, created_at) 
		VALUES (?, ?, NOW())
	`, userID, calendarID)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la création de la liaison user_calendar: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("erreur lors du commit de la transaction: %v", err)
	}

	return int(calendarID), nil
}

// createEventForUser crée un événement dans le premier calendrier de l'utilisateur
func createEventForUser(userID int, title, description string) (int, error) {
	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("erreur lors du démarrage de la transaction: %v", err)
	}
	defer tx.Rollback()

	// Récupérer le premier calendrier de l'utilisateur
	var calendarID int
	err = tx.QueryRow(`
		SELECT uc.calendar_id 
		FROM user_calendar uc
		INNER JOIN calendar c ON uc.calendar_id = c.calendar_id
		WHERE uc.user_id = ? AND uc.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY uc.created_at ASC
		LIMIT 1
	`, userID).Scan(&calendarID)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération du calendrier: %v", err)
	}

	// Créer l'événement
	startTime := time.Now().Add(1 * time.Hour) // Événement dans 1 heure
	result, err := tx.Exec(`
		INSERT INTO event (title, description, start, duration, canceled, created_at) 
		VALUES (?, ?, ?, ?, ?, NOW())
	`, title, description, startTime, 60, false) // Durée de 60 minutes
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la création de l'événement: %v", err)
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération de l'ID de l'événement: %v", err)
	}

	// Associer l'événement au calendrier
	_, err = tx.Exec(`
		INSERT INTO calendar_event (calendar_id, event_id, created_at) 
		VALUES (?, ?, NOW())
	`, calendarID, eventID)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la création de la liaison calendar_event: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("erreur lors du commit de la transaction: %v", err)
	}

	return int(eventID), nil
}
