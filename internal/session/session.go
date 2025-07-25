// Package session internal/session/session.go
package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"go-averroes/internal/common"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SessionStruct struct{}

var Session = SessionStruct{}

// Login authentifie un utilisateur et crée une session
// @Summary Connexion utilisateur
// @Description Authentifie un utilisateur et crée une session. Retourne un token de session, un refresh token, l'utilisateur et ses rôles.
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body common.LoginRequest true "Données de connexion"
// @Success 200 {object} common.JSONResponse
// @Failure 400 {object} common.JSONErrorResponse
// @Failure 401 {object} common.JSONErrorResponse
// @Router /auth/login [post]
func (SessionStruct) Login(c *gin.Context) {
	slog.Info(common.LogLoginAttempt)
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(fmt.Sprintf(common.LogInvalidLoginData, err.Error()))
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	// Récupérer l'utilisateur et son mot de passe
	var user common.User
	var passwordHash string
	err := common.DB.QueryRow(`
		SELECT u.user_id, u.lastname, u.firstname, u.email, u.created_at, u.updated_at, u.deleted_at, up.password_hash
		FROM user u
		INNER JOIN user_password up ON u.user_id = up.user_id
		WHERE u.email = ? AND u.deleted_at IS NULL AND up.deleted_at IS NULL
	`, req.Email).Scan(
		&user.UserID,
		&user.Lastname,
		&user.Firstname,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&passwordHash,
	)

	if err == sql.ErrNoRows {
		slog.Error(common.ErrUserNotFound)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidCredentials,
			Data:    req,
		})
		return
	}

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		slog.Error(common.LogInvalidPassword)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidCredentials,
			Data:    req,
		})
		return
	}

	// Récupérer les rôles de l'utilisateur
	roles, err := GetUserRoles(user.UserID)
	if err != nil {
		slog.Error(fmt.Sprintf(common.LogRolesRetrievalError, err.Error()))
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleNotFound,
		})
		return
	}

	// Générer les tokens
	sessionToken, err := generateToken()
	if err != nil {
		slog.Error("Erreur lors de la génération du token de session: " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTokenGeneration,
		})
		return
	}

	refreshToken, err := generateToken()
	if err != nil {
		slog.Error("Erreur lors de la génération du refresh token: " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTokenGeneration,
		})
		return
	}

	// Définir l'expiration (1 heure pour le session token)
	sessionExpiresAt := time.Now().Add(1 * time.Hour)

	// Récupérer les informations de l'appareil
	deviceInfo := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	// Récupérer la localisation géographique à partir de l'IP
	location := common.GetLocationFromIP(ipAddress)

	// Debug: afficher les informations récupérées
	slog.Info("Informations de session",
		"device_info", deviceInfo,
		"ip_address", ipAddress,
		"location", location,
	)

	// Créer la session en base
	_, err = common.DB.Exec(`
		INSERT INTO user_session (user_id, session_token, refresh_token, expires_at, device_info, ip_address, location, is_active, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, TRUE, NOW())
	`, user.UserID, sessionToken, refreshToken, sessionExpiresAt, deviceInfo, ipAddress, location)
	if err != nil {
		slog.Error("Erreur lors de la création de la session: " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionCreation,
		})
		return
	}

	response := common.LoginResponse{
		User:         user,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    sessionExpiresAt,
		Roles:        roles,
	}

	slog.Info(fmt.Sprintf(common.LogLoginSuccess, user.Email))
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessLogin,
		Data:    response,
	})
}

// Logout déconnecte un utilisateur en supprimant sa session
// @Summary Déconnexion utilisateur
// @Description Déconnecte l'utilisateur en supprimant sa session courante.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} common.JSONResponse
// @Failure 401 {object} common.JSONErrorResponse
// @Router /auth/logout [post]
func (SessionStruct) Logout(c *gin.Context) {
	slog.Info(common.LogLogoutAttempt)

	// Récupérer le token depuis le header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		slog.Error(common.LogMissingAuthHeader)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotAuthenticated,
		})
		return
	}

	// Extraire le token (format: "Bearer <token>")
	token := extractTokenFromHeader(authHeader)
	if token == "" {
		slog.Error(common.LogInvalidToken)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionInvalid,
		})
		return
	}

	// Désactiver la session
	_, err := common.DB.Exec(`
		UPDATE user_session 
		SET is_active = FALSE, updated_at = NOW() 
		WHERE session_token = ? AND is_active = TRUE
	`, token)
	if err != nil {
		slog.Error("Erreur lors de la déconnexion: " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionDeletion,
		})
		return
	}

	slog.Info(common.LogLogoutSuccess)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessLogout,
	})
}

// RefreshToken rafraîchit un token de session
// @Summary Rafraîchissement de token
// @Description Rafraîchit un token de session à partir d'un refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param refreshToken body common.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} common.JSONResponse
// @Failure 400 {object} common.JSONErrorResponse
// @Failure 401 {object} common.JSONErrorResponse
// @Router /auth/refresh [post]
func (SessionStruct) RefreshToken(c *gin.Context) {
	slog.Info(common.LogRefreshTokenAttempt)

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(fmt.Sprintf(common.LogInvalidData, err.Error()))
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	// Vérifier le refresh token
	var session common.UserSession
	err := common.DB.QueryRow(`
		SELECT user_session_id, user_id, session_token, refresh_token, expires_at, device_info, ip_address, location, is_active, created_at, updated_at, deleted_at
		FROM user_session 
		WHERE refresh_token = ? AND is_active = TRUE AND deleted_at IS NULL
	`, req.RefreshToken).Scan(
		&session.UserSessionID,
		&session.UserID,
		&session.SessionToken,
		&session.RefreshToken,
		&session.ExpiresAt,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.Location,
		&session.IsActive,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.DeletedAt,
	)

	if err == sql.ErrNoRows {
		slog.Error(common.LogInvalidRefreshToken)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionInvalid,
		})
		return
	}

	// Vérifier si le refresh token n'est pas expiré
	if time.Now().After(session.ExpiresAt) {
		slog.Error(common.LogRefreshTokenExpired)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionExpired,
		})
		return
	}

	// Générer un nouveau session token
	newSessionToken, err := generateToken()
	if err != nil {
		slog.Error(fmt.Sprintf(common.LogTokenGenerationError, err.Error()))
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTokenGeneration,
		})
		return
	}

	// Mettre à jour la session avec le nouveau token
	newExpiresAt := time.Now().Add(1 * time.Hour)
	_, err = common.DB.Exec(`
		UPDATE user_session 
		SET session_token = ?, expires_at = ?, updated_at = NOW() 
		WHERE user_session_id = ?
	`, newSessionToken, newExpiresAt, session.UserSessionID)
	if err != nil {
		slog.Error(fmt.Sprintf(common.LogSessionUpdateError, err.Error()))
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionUpdate,
		})
		return
	}

	slog.Info(common.LogTokenRefreshSuccess)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessRefreshToken,
		Data: gin.H{
			"session_token": newSessionToken,
			"expires_at":    newExpiresAt,
		},
	})
}

// GetUserSessions récupère toutes les sessions d'un utilisateur
// @Summary Liste des sessions utilisateur
// @Description Récupère toutes les sessions actives de l'utilisateur connecté.
// @Tags Auth
// @Produce json
// @Success 200 {object} common.JSONResponse
// @Failure 401 {object} common.JSONErrorResponse
// @Router /auth/sessions [get]
func (SessionStruct) GetUserSessions(c *gin.Context) {
	slog.Info(common.LogGetUserSessions)

	userData, ok := common.GetUserFromContext(c)
	if !ok {
		slog.Error(common.LogUserNotFoundInContext)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotAuthenticated,
		})
		return
	}

	rows, err := common.DB.Query(`
		SELECT user_session_id, user_id, session_token, refresh_token, expires_at, device_info, ip_address, location, is_active, created_at, updated_at, deleted_at
		FROM user_session 
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userData.UserID)
	if err != nil {
		slog.Error(fmt.Sprintf(common.LogSessionsRetrievalError, err.Error()))
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionNotFound,
		})
		return
	}
	defer rows.Close()

	var sessions []common.UserSession
	for rows.Next() {
		var session common.UserSession
		err := rows.Scan(&session.UserSessionID, &session.UserID, &session.SessionToken, &session.RefreshToken, &session.ExpiresAt, &session.DeviceInfo, &session.IPAddress, &session.Location, &session.IsActive, &session.CreatedAt, &session.UpdatedAt, &session.DeletedAt)
		if err != nil {
			slog.Error(fmt.Sprintf(common.LogSessionReadingError, err.Error()))
			continue
		}
		// Masquer les tokens pour la sécurité
		session.SessionToken = "***"
		session.RefreshToken = nil
		sessions = append(sessions, session)
	}

	slog.Info(common.LogSessionsRetrieved)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    sessions,
	})
}

// DeleteSession supprime une session spécifique
// @Summary Suppression d'une session utilisateur
// @Description Supprime une session spécifique de l'utilisateur connecté.
// @Tags Auth
// @Produce json
// @Param session_id path int true "ID de la session"
// @Success 200 {object} common.JSONResponse
// @Failure 400 {object} common.JSONErrorResponse
// @Failure 401 {object} common.JSONErrorResponse
// @Failure 404 {object} common.JSONErrorResponse
// @Router /auth/sessions/{session_id} [delete]
func (SessionStruct) DeleteSession(c *gin.Context) {
	slog.Info(common.LogDeleteSession)

	userData, ok := common.GetUserFromContext(c)
	if !ok {
		slog.Error(common.LogUserNotFoundInContext)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotAuthenticated,
		})
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		slog.Error(common.LogMissingSessionID)
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	// Vérifier que la session appartient à l'utilisateur
	var existingSessionID int
	err := common.DB.QueryRow("SELECT user_session_id FROM user_session WHERE user_session_id = ? AND user_id = ? AND deleted_at IS NULL", sessionID, userData.UserID).Scan(&existingSessionID)
	if err == sql.ErrNoRows {
		slog.Error(common.LogSessionNotFoundOrNotOwned)
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionNotFound,
		})
		return
	}

	// Supprimer la session
	_, err = common.DB.Exec("UPDATE user_session SET deleted_at = NOW() WHERE user_session_id = ?", sessionID)
	if err != nil {
		slog.Error(fmt.Sprintf(common.LogSessionDeletionError, err.Error()))
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrSessionDeletion,
		})
		return
	}

	slog.Info(common.LogSessionDeletedSuccess)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessDeleteSession,
	})
}

// ValidateSession valide un token de session (utilisé par le middleware)
func (SessionStruct) ValidateSession(token string) (*common.User, error) {
	var user common.User
	var expiresAt time.Time
	err := common.DB.QueryRow(`
		SELECT u.user_id, u.lastname, u.firstname, u.email, u.created_at, u.updated_at, u.deleted_at, us.expires_at
		FROM user u
		INNER JOIN user_session us ON u.user_id = us.user_id
		WHERE us.session_token = ? AND us.is_active = TRUE AND us.deleted_at IS NULL AND u.deleted_at IS NULL
	`, token).Scan(
		&user.UserID,
		&user.Lastname,
		&user.Firstname,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&expiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(common.ErrSessionNotFound)
	}

	if time.Now().After(expiresAt) {
		return nil, errors.New(common.ErrSessionExpired)
	}

	return &user, nil
}

// Fonctions utilitaires

// generateToken génère un token aléatoire
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// extractTokenFromHeader extrait le token du header Authorization
func extractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// GetUserRoles récupère les rôles d'un utilisateur
func GetUserRoles(userID int) ([]common.Role, error) {
	rows, err := common.DB.Query(`
		SELECT r.role_id, r.name, r.description, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		INNER JOIN user_roles ur ON r.role_id = ur.role_id
		WHERE ur.user_id = ? AND ur.deleted_at IS NULL AND r.deleted_at IS NULL
		ORDER BY r.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []common.Role
	for rows.Next() {
		var role common.Role
		err := rows.Scan(
			&role.RoleID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
			&role.DeletedAt,
		)
		if err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}
