// Package user internal/user/user.go
package user

import (
	"database/sql"
	"go-averroes/internal/common"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserStruct struct{}

var User = UserStruct{}

// Get récupère un utilisateur par son ID
func (UserStruct) Get(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID utilisateur invalide",
		})
		return
	}

	var user common.User
	err = common.DB.QueryRow(`
		SELECT user_id, lastname, firstname, email, created_at, updated_at, deleted_at 
		FROM user 
		WHERE user_id = ? AND deleted_at IS NULL
	`, userID).Scan(&user.UserID, &user.Lastname, &user.Firstname, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, common.JSONResponse{
				Success: false,
				Error:   "Utilisateur non trouvé",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la récupération de l'utilisateur",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    user,
	})
}

// Add crée un nouvel utilisateur
func (UserStruct) Add(c *gin.Context) {
	var req common.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

	// Vérifier si l'email existe déjà
	var existingID int
	err := common.DB.QueryRow("SELECT user_id FROM user WHERE email = ? AND deleted_at IS NULL", req.Email).Scan(&existingID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusConflict, common.JSONResponse{
			Success: false,
			Error:   "Un utilisateur avec cet email existe déjà",
		})
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors du hashage du mot de passe",
		})
		return
	}

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors du démarrage de la transaction",
		})
		return
	}
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Insérer l'utilisateur
	result, err := tx.Exec(`
		INSERT INTO user (lastname, firstname, email, created_at) 
		VALUES (?, ?, ?, NOW())
	`, req.Lastname, req.Firstname, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la création de l'utilisateur",
		})
		return
	}

	userID, _ := result.LastInsertId()

	// Insérer le mot de passe
	_, err = tx.Exec(`
		INSERT INTO user_password (user_id, password_hash, created_at) 
		VALUES (?, ?, NOW())
	`, userID, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la création du mot de passe",
		})
		return
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la validation de la transaction",
		})
		return
	}

	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: "Utilisateur créé avec succès",
		Data:    gin.H{"user_id": userID},
	})
}

// Update met à jour un utilisateur
func (UserStruct) Update(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID utilisateur invalide",
		})
		return
	}

	var req common.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "Données invalides: " + err.Error(),
		})
		return
	}

	// Validation des données
	if req.Email != nil {
		// Validation de l'email
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(*req.Email) {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   "Format d'email invalide",
			})
			return
		}
	}

	if req.Password != nil {
		// Validation du mot de passe
		if len(*req.Password) < 6 {
			c.JSON(http.StatusBadRequest, common.JSONResponse{
				Success: false,
				Error:   "Le mot de passe doit contenir au moins 6 caractères",
			})
			return
		}
	}

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors du démarrage de la transaction",
		})
		return
	}
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Vérifier si l'utilisateur existe
	var existingUser common.User
	err = tx.QueryRow("SELECT user_id FROM user WHERE user_id = ? AND deleted_at IS NULL", userID).Scan(&existingUser.UserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Utilisateur non trouvé",
		})
		return
	}

	// Construire la requête de mise à jour
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
		// Vérifier si le nouvel email existe déjà
		var existingID int
		err = tx.QueryRow("SELECT user_id FROM user WHERE email = ? AND user_id != ? AND deleted_at IS NULL", *req.Email, userID).Scan(&existingID)
		if err != sql.ErrNoRows {
			c.JSON(http.StatusConflict, common.JSONResponse{
				Success: false,
				Error:   "Un utilisateur avec cet email existe déjà",
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
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la mise à jour de l'utilisateur",
		})
		return
	}

	// Mettre à jour le mot de passe si fourni
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   "Erreur lors du hashage du mot de passe",
			})
			return
		}

		_, err = tx.Exec(`
			UPDATE user_password 
			SET password_hash = ?, updated_at = NOW() 
			WHERE user_id = ?
		`, string(hashedPassword), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.JSONResponse{
				Success: false,
				Error:   "Erreur lors de la mise à jour du mot de passe",
			})
			return
		}
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la validation de la transaction",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: "Utilisateur mis à jour avec succès",
	})
}

// Delete supprime un utilisateur (soft delete)
func (UserStruct) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   "ID utilisateur invalide",
		})
		return
	}

	// Démarrer une transaction
	tx, err := common.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors du démarrage de la transaction",
		})
		return
	}
	defer tx.Rollback() // Rollback par défaut, commit seulement si tout va bien

	// Vérifier si l'utilisateur existe
	var existingUser common.User
	err = tx.QueryRow("SELECT user_id FROM user WHERE user_id = ? AND deleted_at IS NULL", userID).Scan(&existingUser.UserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Utilisateur non trouvé",
		})
		return
	}

	// Soft delete de l'utilisateur
	_, err = tx.Exec("UPDATE user SET deleted_at = NOW() WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression de l'utilisateur",
		})
		return
	}

	// Soft delete du mot de passe
	_, err = tx.Exec("UPDATE user_password SET deleted_at = NOW() WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la suppression du mot de passe",
		})
		return
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   "Erreur lors de la validation de la transaction",
		})
		return
	}

	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: "Utilisateur supprimé avec succès",
	})
}
