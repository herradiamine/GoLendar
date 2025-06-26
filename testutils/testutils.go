package testutils

import (
	"fmt"
	"os"
	"time"

	"go-averroes/internal/common"

	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // Driver MySQL
	"golang.org/x/crypto/bcrypt"
)

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

// PurgeTestData supprime toutes les données de test d'un utilisateur par son email
// Supprime dans l'ordre correct pour respecter les contraintes de clés étrangères
func PurgeTestData(email string) error {
	if email == "" {
		return nil // Pas de données à purger
	}

	// Vérifier que la base de données est initialisée
	if common.DB == nil {
		return fmt.Errorf("base de données non initialisée")
	}

	// Début de transaction pour la purge
	tx, err := common.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur lors du démarrage de la transaction de purge: %v", err)
	}
	defer tx.Rollback()

	// Supprimer les sessions de l'utilisateur
	_, err = tx.Exec("DELETE FROM user_session WHERE user_id IN (SELECT user_id FROM user WHERE email = ?)", email)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des sessions: %v", err)
	}

	// Supprimer les rôles de l'utilisateur
	_, err = tx.Exec("DELETE FROM user_roles WHERE user_id IN (SELECT user_id FROM user WHERE email = ?)", email)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des rôles: %v", err)
	}

	// Supprimer les mots de passe de l'utilisateur
	_, err = tx.Exec("DELETE FROM user_password WHERE user_id IN (SELECT user_id FROM user WHERE email = ?)", email)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des mots de passe: %v", err)
	}

	// Supprimer l'utilisateur
	_, err = tx.Exec("DELETE FROM user WHERE email = ?", email)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'utilisateur: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur lors de la validation de la transaction de purge: %v", err)
	}

	return nil
}

// PurgeTestUser alias de PurgeTestData pour la compatibilité
// Cette fonction est maintenue pour la compatibilité avec l'ancien code
func PurgeTestUser(email string) error {
	return PurgeTestData(email)
}

// PurgeTestCalendar supprime un calendrier de test et toutes ses données associées
func PurgeTestCalendar(calendarID int) error {
	if calendarID <= 0 {
		return nil // Pas de données à purger
	}

	// Vérifier que la base de données est initialisée
	if common.DB == nil {
		return fmt.Errorf("base de données non initialisée")
	}

	// Début de transaction pour la purge
	tx, err := common.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur lors du démarrage de la transaction de purge calendrier: %v", err)
	}
	defer tx.Rollback()

	// Supprimer les événements du calendrier (via calendar_event)
	_, err = tx.Exec("DELETE FROM calendar_event WHERE calendar_id = ?", calendarID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des événements du calendrier: %v", err)
	}

	// Supprimer les liaisons utilisateur-calendrier
	_, err = tx.Exec("DELETE FROM user_calendar WHERE calendar_id = ?", calendarID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des liaisons utilisateur-calendrier: %v", err)
	}

	// Supprimer le calendrier
	_, err = tx.Exec("DELETE FROM calendar WHERE calendar_id = ?", calendarID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du calendrier: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur lors de la validation de la transaction de purge calendrier: %v", err)
	}

	return nil
}

// PurgeTestEvent supprime un événement de test et toutes ses données associées
func PurgeTestEvent(eventID int) error {
	if eventID <= 0 {
		return nil // Pas de données à purger
	}

	// Vérifier que la base de données est initialisée
	if common.DB == nil {
		return fmt.Errorf("base de données non initialisée")
	}

	// Début de transaction pour la purge
	tx, err := common.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur lors du démarrage de la transaction de purge événement: %v", err)
	}
	defer tx.Rollback()

	// Supprimer les liaisons calendrier-événement
	_, err = tx.Exec("DELETE FROM calendar_event WHERE event_id = ?", eventID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des liaisons calendrier-événement: %v", err)
	}

	// Supprimer l'événement
	_, err = tx.Exec("DELETE FROM event WHERE event_id = ?", eventID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'événement: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur lors de la validation de la transaction de purge événement: %v", err)
	}

	return nil
}

// PurgeTestRole supprime un rôle de test et toutes ses données associées
func PurgeTestRole(roleID int) error {
	if roleID <= 0 {
		return nil // Pas de données à purger
	}

	// Vérifier que la base de données est initialisée
	if common.DB == nil {
		return fmt.Errorf("base de données non initialisée")
	}

	// Début de transaction pour la purge
	tx, err := common.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur lors du démarrage de la transaction de purge rôle: %v", err)
	}
	defer tx.Rollback()

	// Supprimer les attributions de rôles
	_, err = tx.Exec("DELETE FROM user_roles WHERE role_id = ?", roleID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression des attributions de rôles: %v", err)
	}

	// Supprimer le rôle
	_, err = tx.Exec("DELETE FROM roles WHERE role_id = ?", roleID)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression du rôle: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur lors de la validation de la transaction de purge rôle: %v", err)
	}

	return nil
}

// GenerateUniqueName génère un nom unique pour les tests
func GenerateUniqueName(baseName string) string {
	return fmt.Sprintf("%s_%d", baseName, time.Now().UnixNano())
}

// GenerateUniqueTitle génère un titre unique pour les tests
func GenerateUniqueTitle(baseTitle string) string {
	return fmt.Sprintf("%s_%d", baseTitle, time.Now().UnixNano())
}

// SetupAuthContext configure le contexte d'authentification pour les tests
// Utilise la clé "auth_user" qui est la clé standard dans l'application
func SetupAuthContext(c *gin.Context, user common.User) {
	c.Set("auth_user", user)
}

// SetupEmptyContext configure un contexte vide (sans utilisateur authentifié)
// Utile pour tester les cas où l'utilisateur n'est pas authentifié
func SetupEmptyContext(c *gin.Context) {
	// Ne rien faire - le contexte reste vide
}

// CreateTestUser crée un utilisateur de test avec des données par défaut
func CreateTestUser(userID int, email string) common.User {
	return common.User{
		UserID:    userID,
		Lastname:  "Test",
		Firstname: "User",
		Email:     email,
		CreatedAt: time.Now(),
	}
}

// CreateTestUserWithCustomData crée un utilisateur de test avec des données personnalisées
func CreateTestUserWithCustomData(userID int, lastname, firstname, email string) common.User {
	return common.User{
		UserID:    userID,
		Lastname:  lastname,
		Firstname: firstname,
		Email:     email,
		CreatedAt: time.Now(),
	}
}

// CreateUnauthenticatedUser crée un utilisateur de test qui existe dans le contexte mais n'a pas de session valide
// Cet utilisateur sera considéré comme non authentifié par les handlers
func CreateUnauthenticatedUser(userID int, lastname, firstname, email string) common.User {
	return common.User{
		UserID:    userID,
		Lastname:  lastname,
		Firstname: firstname,
		Email:     email,
		CreatedAt: time.Now(),
		// Pas de session associée = non authentifié
	}
}

// CreateAuthenticatedUser crée un utilisateur avec une session valide pour les tests
func CreateAuthenticatedUser(userID int, lastname, firstname, email string) (*common.User, string, error) {
	uniqueUserID := int(time.Now().UnixNano() % 1000000)
	user := common.User{
		UserID:    uniqueUserID,
		Lastname:  lastname,
		Firstname: firstname,
		Email:     email,
		CreatedAt: time.Now(),
	}

	sessionToken, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la génération du token: %v", err)
	}

	refreshToken, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la génération du refresh token: %v", err)
	}

	tx, err := common.DB.Begin()
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors du démarrage de la transaction: %v", err)
	}

	// Insérer l'utilisateur
	_, err = tx.Exec(`
		INSERT INTO user (user_id, lastname, firstname, email, created_at) 
		VALUES (?, ?, ?, ?, NOW())
	`, user.UserID, user.Lastname, user.Firstname, user.Email)
	if err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("erreur lors de la création de l'utilisateur: %v", err)
	}

	// Insérer la session
	_, err = tx.Exec(`
		INSERT INTO user_session (user_id, session_token, refresh_token, expires_at, device_info, ip_address, is_active, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, TRUE, NOW())
	`, user.UserID, sessionToken, refreshToken, time.Now().Add(1*time.Hour), "test-device", "127.0.0.1")
	if err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("erreur lors de la création de la session: %v", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, "", fmt.Errorf("erreur lors du commit de la transaction: %v", err)
	}

	return &user, sessionToken, nil
}

// generateToken génère un token aléatoire (copie de la fonction dans session.go)
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateUserWithPassword crée un utilisateur avec un mot de passe hashé
func CreateUserWithPassword(lastname, firstname, email, password string) (*common.User, error) {
	// Vérifier que la base de données est initialisée
	if common.DB == nil {
		return nil, fmt.Errorf("base de données non initialisée")
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du hash du mot de passe: %v", err)
	}

	// Début de transaction
	tx, err := common.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erreur lors du démarrage de la transaction: %v", err)
	}
	defer tx.Rollback()

	// Insérer l'utilisateur
	result, err := tx.Exec(`
		INSERT INTO user (lastname, firstname, email, created_at) 
		VALUES (?, ?, ?, NOW())
	`, lastname, firstname, email)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de l'utilisateur: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'ID utilisateur: %v", err)
	}

	// Insérer le mot de passe
	_, err = tx.Exec(`
		INSERT INTO user_password (user_id, password_hash, created_at) 
		VALUES (?, ?, NOW())
	`, userID, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du mot de passe: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erreur lors de la validation de la transaction: %v", err)
	}

	// Créer l'objet utilisateur
	user := &common.User{
		UserID:    int(userID),
		Lastname:  lastname,
		Firstname: firstname,
		Email:     email,
		CreatedAt: time.Now(),
	}

	return user, nil
}

// CreateAdminUser crée un utilisateur avec le rôle admin
func CreateAdminUser(userID int, lastname, firstname, email string) (*common.User, string, error) {
	// Vérifier que la base de données est initialisée
	if common.DB == nil {
		return nil, "", fmt.Errorf("base de données non initialisée")
	}

	// Début de transaction
	tx, err := common.DB.Begin()
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors du démarrage de la transaction: %v", err)
	}
	defer tx.Rollback()

	// Insérer l'utilisateur
	result, err := tx.Exec(`
		INSERT INTO user (user_id, lastname, firstname, email, created_at) 
		VALUES (?, ?, ?, ?, NOW())
	`, userID, lastname, firstname, email)
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la création de l'utilisateur: %v", err)
	}

	// Vérifier que l'utilisateur a été créé
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return nil, "", fmt.Errorf("aucun utilisateur créé")
	}

	// Insérer un mot de passe par défaut
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors du hash du mot de passe: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO user_password (user_id, password_hash, created_at) 
		VALUES (?, ?, NOW())
	`, userID, string(hashedPassword))
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la création du mot de passe: %v", err)
	}

	// Vérifier si le rôle admin existe, sinon le créer
	var adminRoleID int
	err = tx.QueryRow("SELECT role_id FROM roles WHERE name = 'admin' AND deleted_at IS NULL").Scan(&adminRoleID)
	if err != nil {
		// Le rôle admin n'existe pas, le créer
		roleResult, err := tx.Exec(`
			INSERT INTO roles (name, description, created_at) 
			VALUES ('admin', 'Administrateur du système', NOW())
		`)
		if err != nil {
			return nil, "", fmt.Errorf("erreur lors de la création du rôle admin: %v", err)
		}
		adminRoleID64, err := roleResult.LastInsertId()
		if err != nil {
			return nil, "", fmt.Errorf("erreur lors de la récupération de l'ID du rôle admin: %v", err)
		}
		adminRoleID = int(adminRoleID64)
	}

	// Assigner le rôle admin à l'utilisateur
	_, err = tx.Exec(`
		INSERT INTO user_roles (user_id, role_id, created_at) 
		VALUES (?, ?, NOW())
	`, userID, adminRoleID)
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de l'assignation du rôle admin: %v", err)
	}

	// Créer une session pour l'utilisateur
	token, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la génération du token: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO user_session (user_id, session_token, expires_at, is_active, created_at) 
		VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 24 HOUR), TRUE, NOW())
	`, userID, token)
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la création de la session: %v", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("erreur lors de la validation de la transaction: %v", err)
	}

	// Créer l'objet utilisateur
	user := &common.User{
		UserID:    userID,
		Lastname:  lastname,
		Firstname: firstname,
		Email:     email,
		CreatedAt: time.Now(),
	}

	return user, token, nil
}

// GetSessionIDForUser récupère le session_token de la session active d'un utilisateur
func GetSessionIDForUser(userID int) (string, error) {
	if common.DB == nil {
		return "", fmt.Errorf("base de données non initialisée")
	}
	var sessionID string
	err := common.DB.QueryRow("SELECT session_token FROM user_session WHERE user_id = ? AND is_active = TRUE LIMIT 1", userID).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("aucune session active trouvée pour l'utilisateur %d: %v", userID, err)
	}
	return sessionID, nil
}

// GetUserSessionIDForUser récupère le user_session_id actif d'un utilisateur
func GetUserSessionIDForUser(userID int) (string, error) {
	if common.DB == nil {
		return "", fmt.Errorf("base de données non initialisée")
	}
	var sessionID string
	err := common.DB.QueryRow("SELECT user_session_id FROM user_session WHERE user_id = ? AND is_active = TRUE AND deleted_at IS NULL LIMIT 1", userID).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("aucune session active trouvée pour l'utilisateur %d: %v", userID, err)
	}
	return sessionID, nil
}

// DeleteUserSessionByID supprime (soft delete) une session par son user_session_id
func DeleteUserSessionByID(sessionID string) error {
	if common.DB == nil {
		return fmt.Errorf("base de données non initialisée")
	}
	_, err := common.DB.Exec("UPDATE user_session SET deleted_at = NOW() WHERE user_session_id = ?", sessionID)
	return err
}

// CreateSessionForUser crée une nouvelle session pour un utilisateur et retourne le token
func CreateSessionForUser(userID int) (string, error) {
	if common.DB == nil {
		return "", fmt.Errorf("base de données non initialisée")
	}
	sessionToken, err := generateToken()
	if err != nil {
		return "", err
	}
	refreshToken, err := generateToken()
	if err != nil {
		return "", err
	}
	_, err = common.DB.Exec(`
		INSERT INTO user_session (user_id, session_token, refresh_token, expires_at, device_info, ip_address, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, TRUE, NOW())
	`, userID, sessionToken, refreshToken, time.Now().Add(1*time.Hour), "test-device-2", "127.0.0.2")
	if err != nil {
		return "", err
	}
	return sessionToken, nil
}

// GetUserSessionIDByToken récupère le user_session_id à partir d'un token
func GetUserSessionIDByToken(token string) (string, error) {
	if common.DB == nil {
		return "", fmt.Errorf("base de données non initialisée")
	}
	var sessionID string
	err := common.DB.QueryRow("SELECT user_session_id FROM user_session WHERE session_token = ? AND deleted_at IS NULL LIMIT 1", token).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("aucune session trouvée pour le token: %v", err)
	}
	return sessionID, nil
}

// GetRefreshTokenForUser récupère le refresh_token actif d'un utilisateur
func GetRefreshTokenForUser(userID int) (string, error) {
	if common.DB == nil {
		return "", fmt.Errorf("base de données non initialisée")
	}
	var refreshToken string
	err := common.DB.QueryRow("SELECT refresh_token FROM user_session WHERE user_id = ? AND is_active = TRUE AND deleted_at IS NULL LIMIT 1", userID).Scan(&refreshToken)
	if err != nil {
		return "", fmt.Errorf("aucun refresh_token actif trouvé pour l'utilisateur %d: %v", userID, err)
	}
	return refreshToken, nil
}

// ExpireRefreshToken rend un refresh_token expiré en base
func ExpireRefreshToken(refreshToken string) error {
	if common.DB == nil {
		return fmt.Errorf("base de données non initialisée")
	}
	_, err := common.DB.Exec("UPDATE user_session SET expires_at = ? WHERE refresh_token = ?", time.Now().Add(-2*time.Hour), refreshToken)
	return err
}
