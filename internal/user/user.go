// Package user internal/user/user.go
package user

import (
	"database/sql"
	"go-averroes/internal/common"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserStruct struct{}

var User = UserStruct{}

// Get récupère un utilisateur par son ID
func (UserStruct) Get(c *gin.Context) {
	slog.Info(common.LogUserGet)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		slog.Error(common.LogUserGet + " - utilisateur non trouvé dans le contexte")
		return
	}

	slog.Info(common.LogUserGet + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    userData,
	})
}

// Add crée un nouvel utilisateur
func (UserStruct) Add(c *gin.Context) {
	slog.Info(common.LogUserAdd)
	var req common.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(common.LogUserAdd + " - données invalides : " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	// Vérifier si l'email existe déjà
	var existingID int
	err := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ? AND deleted_at IS NULL", req.Email).Scan(&existingID)
	if err != sql.ErrNoRows {
		slog.Error(common.LogUserAdd + " - utilisateur déjà existant")
		c.JSON(http.StatusConflict, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserAlreadyExists,
		})
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error(common.LogUserAdd + " - erreur lors du hash du mot de passe : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrPasswordHashing,
		})
		return
	}

	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogUserAdd + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO user (lastname, firstname, email, created_at) 
		VALUES (?, ?, ?, NOW())
	`, req.Lastname, req.Firstname, req.Email)
	if err != nil {
		slog.Error(common.LogUserAdd + " - erreur lors de la création de l'utilisateur : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserCreation,
		})
		return
	}

	userID, _ := result.LastInsertId()

	_, err = tx.Exec(`
		INSERT INTO user_password (user_id, password_hash, created_at) 
		VALUES (?, ?, NOW())
	`, userID, string(hashedPassword))
	if err != nil {
		slog.Error(common.LogUserAdd + " - erreur lors de la création du mot de passe : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrPasswordCreation,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error(common.LogUserAdd + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogUserAdd + " - succès")
	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessCreateUser,
		Data:    gin.H{"user_id": userID},
	})
}

// Update met à jour un utilisateur
func (UserStruct) Update(c *gin.Context) {
	slog.Info(common.LogUserUpdate)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		slog.Error(common.LogUserUpdate + " - utilisateur non trouvé dans le contexte")
		return
	}
	userID := userData.UserID

	var req common.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(common.LogUserUpdate + " - données invalides : " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	if req.Email != nil {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(*req.Email) {
			slog.Error(common.LogUserUpdate + " - format email invalide")
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   common.ErrInvalidEmailFormat,
			})
			return
		}
	}

	if req.Password != nil {
		if len(*req.Password) < 6 {
			slog.Error(common.LogUserUpdate + " - mot de passe trop court")
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   common.ErrPasswordTooShort,
			})
			return
		}
	}

	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogUserUpdate + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	query := "UPDATE user SET updated_at = NOW()"
	var args []interface{}

	if req.Lastname != nil {
		query += ", lastname = ?"
		args = append(args, *req.Lastname)
	}
	if req.Firstname != nil {
		query += ", firstname = ?"
		args = append(args, *req.Firstname)
	}
	if req.Email != nil {
		var existingID int
		err = tx.QueryRow("SELECT user_id FROM user WHERE email = ? AND user_id != ? AND deleted_at IS NULL", *req.Email, userID).Scan(&existingID)
		if err != sql.ErrNoRows {
			slog.Error(common.LogUserUpdate + " - email déjà utilisé")
			c.JSON(http.StatusConflict, common.JSONResponse{
				Success: false,
				Error:   common.ErrUserAlreadyExists,
			})
			return
		}
		query += ", email = ?"
		args = append(args, *req.Email)
	}

	query += " WHERE user_id = ?"
	args = append(args, userID)

	_, err = tx.Exec(query, args...)
	if err != nil {
		slog.Error(common.LogUserUpdate + " - erreur lors de la mise à jour de l'utilisateur : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserUpdate,
		})
		return
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error(common.LogUserUpdate + " - erreur lors du hash du mot de passe : " + err.Error())
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   common.ErrPasswordHashing,
			})
			return
		}

		_, err = tx.Exec(`
			UPDATE user_password 
			SET password_hash = ?, updated_at = NOW() 
			WHERE user_id = ?
		`, string(hashedPassword), userID)
		if err != nil {
			slog.Error(common.LogUserUpdate + " - erreur lors de la mise à jour du mot de passe : " + err.Error())
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   common.ErrPasswordUpdate,
			})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error(common.LogUserUpdate + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogUserUpdate + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUserUpdate,
	})
}

// Delete supprime un utilisateur (soft delete)
func (UserStruct) Delete(c *gin.Context) {
	slog.Info(common.LogUserDelete)
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		slog.Error(common.LogUserDelete + " - utilisateur non trouvé dans le contexte")
		return
	}
	userID := userData.UserID

	tx, err := common.DB.Begin()
	if err != nil {
		slog.Error(common.LogUserDelete + " - erreur lors du démarrage de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionStart,
		})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE user SET deleted_at = NOW() WHERE user_id = ?", userID)
	if err != nil {
		slog.Error(common.LogUserDelete + " - erreur lors de la suppression de l'utilisateur : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserDelete,
		})
		return
	}

	_, err = tx.Exec("UPDATE user_password SET deleted_at = NOW() WHERE user_id = ?", userID)
	if err != nil {
		slog.Error(common.LogUserDelete + " - erreur lors de la suppression du mot de passe : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrPasswordDelete,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.Error(common.LogUserDelete + " - erreur lors du commit de la transaction : " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrTransactionCommit,
		})
		return
	}

	slog.Info(common.LogUserDelete + " - succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUserDelete,
	})
}

// GetUserWithRoles récupère un utilisateur avec ses rôles
func (UserStruct) GetUserWithRoles(c *gin.Context) {
	slog.Info("Récupération d'un utilisateur avec ses rôles")
	userData, ok := common.GetUserFromContext(c)
	if !ok {
		slog.Error(common.LogUserNotFoundInContext)
		c.JSON(http.StatusUnauthorized, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotAuthenticated,
		})
		return
	}

	// Récupérer les rôles de l'utilisateur
	rows, err := common.DB.Query(`
		SELECT r.role_id, r.name, r.description, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		INNER JOIN user_roles ur ON r.role_id = ur.role_id
		WHERE ur.user_id = ? AND ur.deleted_at IS NULL AND r.deleted_at IS NULL
		ORDER BY r.name
	`, userData.UserID)
	if err != nil {
		slog.Error("Erreur lors de la récupération des rôles: " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleNotFound,
		})
		return
	}
	defer rows.Close()

	var roles []common.Role
	for rows.Next() {
		var role common.Role
		err := rows.Scan(&role.RoleID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt)
		if err != nil {
			slog.Error("Erreur lors de la lecture du rôle: " + err.Error())
			continue
		}
		roles = append(roles, role)
	}

	userWithRoles := common.UserWithRoles{
		User:  userData,
		Roles: roles,
	}

	slog.Info("Utilisateur avec rôles récupéré avec succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    userWithRoles,
	})
}
