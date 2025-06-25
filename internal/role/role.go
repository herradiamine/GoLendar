// Package role internal/role/role.go
package role

import (
	"database/sql"
	"go-averroes/internal/common"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoleStruct struct{}

var Role = RoleStruct{}

// Get récupère un rôle par son ID
func (RoleStruct) Get(c *gin.Context) {
	slog.Info("Récupération d'un rôle")
	roleID := c.Param("id")
	if roleID == "" {
		slog.Error(common.ErrMissingRoleID)
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	var role common.Role
	err := common.DB.QueryRow(`
		SELECT role_id, name, description, created_at, updated_at, deleted_at
		FROM roles 
		WHERE role_id = ? AND deleted_at IS NULL
	`, roleID).Scan(&role.RoleID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt)

	if err == sql.ErrNoRows {
		slog.Error(common.ErrRoleNotFound)
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleNotFound,
		})
		return
	}

	if err != nil {
		slog.Error("Erreur lors de la récupération du rôle: " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleNotFound,
		})
		return
	}

	slog.Info("Rôle récupéré avec succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    role,
	})
}

// List récupère tous les rôles
func (RoleStruct) List(c *gin.Context) {
	slog.Info("Récupération de tous les rôles")

	rows, err := common.DB.Query(`
		SELECT role_id, name, description, created_at, updated_at, deleted_at 
		FROM roles 
		WHERE deleted_at IS NULL 
		ORDER BY name
	`)
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

	slog.Info("Rôles récupérés avec succès")
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    roles,
	})
}

// Add crée un nouveau rôle
func (RoleStruct) Add(c *gin.Context) {
	slog.Info("Création d'un nouveau rôle")
	var req common.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Données invalides: " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	// Vérifier si le nom du rôle existe déjà
	var existingID int
	err := common.DB.QueryRow("SELECT role_id FROM roles WHERE name = ? AND deleted_at IS NULL", req.Name).Scan(&existingID)
	if err != sql.ErrNoRows {
		slog.Error("Rôle déjà existant")
		c.JSON(http.StatusConflict, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleAlreadyExists,
		})
		return
	}

	result, err := common.DB.Exec(`
		INSERT INTO roles (name, description, created_at) 
		VALUES (?, ?, NOW())
	`, req.Name, req.Description)
	if err != nil {
		slog.Error("Erreur lors de la création du rôle: " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleCreation,
		})
		return
	}

	roleID, _ := result.LastInsertId()

	slog.Info(common.LogRoleCreateSuccess)
	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessCreateRole,
		Data:    gin.H{"role_id": roleID},
	})
}

// Update met à jour un rôle
func (RoleStruct) Update(c *gin.Context) {
	slog.Info("Mise à jour d'un rôle")
	roleID := c.Param("id")
	if roleID == "" {
		slog.Error(common.ErrMissingRoleID)
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	var req common.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Données invalides: " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	// Vérifier si le rôle existe
	var existingRole common.Role
	err := common.DB.QueryRow("SELECT role_id FROM roles WHERE role_id = ? AND deleted_at IS NULL", roleID).Scan(&existingRole.RoleID)
	if err == sql.ErrNoRows {
		slog.Error(common.ErrRoleNotFound)
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleNotFound,
		})
		return
	}

	// Vérifier si le nouveau nom existe déjà (si fourni)
	if req.Name != nil {
		var existingID int
		err = common.DB.QueryRow("SELECT role_id FROM roles WHERE name = ? AND role_id != ? AND deleted_at IS NULL", *req.Name, roleID).Scan(&existingID)
		if err != sql.ErrNoRows {
			slog.Error(common.ErrRoleNameAlreadyUsed)
			c.JSON(http.StatusConflict, common.JSONResponse{
				Success: false,
				Error:   common.ErrRoleAlreadyExists,
			})
			return
		}
	}

	query := "UPDATE roles SET updated_at = NOW()"
	var args []interface{}

	if req.Name != nil {
		query += ", name = ?"
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		query += ", description = ?"
		args = append(args, *req.Description)
	}

	args = append(args, roleID)
	_, err = common.DB.Exec(query+" WHERE role_id = ?", args...)
	if err != nil {
		slog.Error(common.ErrRoleUpdateFailed + ": " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleUpdate,
		})
		return
	}

	slog.Info(common.LogRoleUpdateSuccess)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessUpdateRole,
	})
}

// Delete supprime un rôle
func (RoleStruct) Delete(c *gin.Context) {
	slog.Info("Suppression d'un rôle")
	roleID := c.Param("id")
	if roleID == "" {
		slog.Error(common.ErrMissingRoleID)
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	// Vérifier si le rôle existe
	var existingRole common.Role
	err := common.DB.QueryRow("SELECT role_id FROM roles WHERE role_id = ? AND deleted_at IS NULL", roleID).Scan(&existingRole.RoleID)
	if err == sql.ErrNoRows {
		slog.Error(common.ErrRoleNotFound)
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleNotFound,
		})
		return
	}

	// Soft delete du rôle
	_, err = common.DB.Exec("UPDATE roles SET deleted_at = NOW() WHERE role_id = ?", roleID)
	if err != nil {
		slog.Error(common.ErrRoleDeleteFailed + ": " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleDelete,
		})
		return
	}

	// Supprimer aussi les attributions de rôles
	_, err = common.DB.Exec("UPDATE user_roles SET deleted_at = NOW() WHERE role_id = ?", roleID)
	if err != nil {
		slog.Error("Erreur lors de la suppression des attributions de rôles: " + err.Error())
		// On continue quand même car le rôle a été supprimé
	}

	slog.Info(common.LogRoleDeleteSuccess)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessDeleteRole,
	})
}

// AssignRole attribue un rôle à un utilisateur
func (RoleStruct) AssignRole(c *gin.Context) {
	slog.Info("Attribution d'un rôle à un utilisateur")
	var req common.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Données invalides: " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	// Vérifier si l'utilisateur existe
	var userID int
	err := common.DB.QueryRow("SELECT user_id FROM user WHERE user_id = ? AND deleted_at IS NULL", req.UserID).Scan(&userID)
	if err == sql.ErrNoRows {
		slog.Error(common.ErrUserNotFound)
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotFound,
		})
		return
	}

	// Vérifier si le rôle existe
	var roleID int
	err = common.DB.QueryRow("SELECT role_id FROM roles WHERE role_id = ? AND deleted_at IS NULL", req.RoleID).Scan(&roleID)
	if err == sql.ErrNoRows {
		slog.Error(common.ErrRoleNotFound)
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleNotFound,
		})
		return
	}

	// Vérifier si l'attribution existe déjà
	var existingID int
	err = common.DB.QueryRow("SELECT user_roles_id FROM user_roles WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL", req.UserID, req.RoleID).Scan(&existingID)
	if err != sql.ErrNoRows {
		slog.Error(common.ErrRoleAttributionConflict)
		c.JSON(http.StatusConflict, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleAlreadyAssigned,
		})
		return
	}

	// Créer l'attribution
	_, err = common.DB.Exec(`
		INSERT INTO user_roles (user_id, role_id, created_at) 
		VALUES (?, ?, NOW())
	`, req.UserID, req.RoleID)
	if err != nil {
		slog.Error(common.ErrRoleAssignmentFailed + ": " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleAssignment,
		})
		return
	}

	slog.Info(common.LogRoleAssignSuccess)
	c.JSON(http.StatusCreated, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessAssignRole,
	})
}

// RevokeRole révoque un rôle d'un utilisateur
func (RoleStruct) RevokeRole(c *gin.Context) {
	slog.Info("Révocation d'un rôle d'un utilisateur")
	var req common.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Données invalides: " + err.Error())
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData + ": " + err.Error(),
		})
		return
	}

	// Vérifier si l'attribution existe
	var userRoleID int
	err := common.DB.QueryRow("SELECT user_roles_id FROM user_roles WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL", req.UserID, req.RoleID).Scan(&userRoleID)
	if err == sql.ErrNoRows {
		slog.Error("Attribution de rôle non trouvée")
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   "Ce rôle n'est pas attribué à cet utilisateur",
		})
		return
	}

	// Supprimer l'attribution
	_, err = common.DB.Exec("UPDATE user_roles SET deleted_at = NOW() WHERE user_roles_id = ?", userRoleID)
	if err != nil {
		slog.Error(common.ErrRoleRevocationFailed + ": " + err.Error())
		c.JSON(http.StatusInternalServerError, common.JSONResponse{
			Success: false,
			Error:   common.ErrRoleRevocation,
		})
		return
	}

	slog.Info(common.LogRoleRevokeSuccess)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Message: common.MsgSuccessRevokeRole,
	})
}

// GetUserRoles récupère les rôles d'un utilisateur
func (RoleStruct) GetUserRoles(c *gin.Context) {
	slog.Info("Récupération des rôles d'un utilisateur")
	userID := c.Param("user_id")
	if userID == "" {
		slog.Error("ID utilisateur manquant")
		c.JSON(http.StatusBadRequest, common.JSONResponse{
			Success: false,
			Error:   common.ErrInvalidData,
		})
		return
	}

	// Vérifier si l'utilisateur existe
	var existingUserID int
	err := common.DB.QueryRow("SELECT user_id FROM user WHERE user_id = ? AND deleted_at IS NULL", userID).Scan(&existingUserID)
	if err == sql.ErrNoRows {
		slog.Error(common.ErrUserNotFound)
		c.JSON(http.StatusNotFound, common.JSONResponse{
			Success: false,
			Error:   common.ErrUserNotFound,
		})
		return
	}

	rows, err := common.DB.Query(`
		SELECT r.role_id, r.name, r.description, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		INNER JOIN user_roles ur ON r.role_id = ur.role_id
		WHERE ur.user_id = ? AND ur.deleted_at IS NULL AND r.deleted_at IS NULL
		ORDER BY r.name
	`, userID)
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

	slog.Info(common.LogUserRolesRetrieved)
	c.JSON(http.StatusOK, common.JSONResponse{
		Success: true,
		Data:    roles,
	})
}
