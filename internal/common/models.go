package common

import "time"

// User représente la table user
type User struct {
	UserID    int        `json:"user_id" db:"user_id"`
	Lastname  string     `json:"lastname" db:"lastname"`
	Firstname string     `json:"firstname" db:"firstname"`
	Email     string     `json:"email" db:"email"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserPassword représente la table user_password
type UserPassword struct {
	UserPasswordID int        `json:"user_password_id" db:"user_password_id"`
	UserID         int        `json:"user_id" db:"user_id"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	ExpiredAt      *time.Time `json:"expired_at,omitempty" db:"expired_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Calendar représente la table calendar
type Calendar struct {
	CalendarID  int        `json:"calendar_id" db:"calendar_id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Event représente la table event
type Event struct {
	EventID     int        `json:"event_id" db:"event_id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	Start       time.Time  `json:"start" db:"start"`
	Duration    int        `json:"duration" db:"duration"`
	Canceled    bool       `json:"canceled" db:"canceled"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserCalendar représente la table user_calendar
type UserCalendar struct {
	UserCalendarID int        `json:"user_calendar_id" db:"user_calendar_id"`
	UserID         int        `json:"user_id" db:"user_id"`
	CalendarID     int        `json:"calendar_id" db:"calendar_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CalendarEvent représente la table calendar_event
type CalendarEvent struct {
	CalendarEventID int        `json:"calendar_event_id" db:"calendar_event_id"`
	CalendarID      int        `json:"calendar_id" db:"calendar_id"`
	EventID         int        `json:"event_id" db:"event_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserCalendarWithDetails représente un calendrier avec ses détails pour un utilisateur
type UserCalendarWithDetails struct {
	UserCalendarID int        `json:"user_calendar_id" db:"user_calendar_id"`
	UserID         int        `json:"user_id" db:"user_id"`
	CalendarID     int        `json:"calendar_id" db:"calendar_id"`
	Title          string     `json:"title" db:"title"`
	Description    *string    `json:"description,omitempty" db:"description"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Role représente la table roles
type Role struct {
	RoleID      int        `json:"role_id" db:"role_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserRole représente la table user_roles
type UserRole struct {
	UserRolesID int        `json:"user_roles_id" db:"user_roles_id"`
	UserID      int        `json:"user_id" db:"user_id"`
	RoleID      int        `json:"role_id" db:"role_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserSession représente la table user_session
type UserSession struct {
	UserSessionID int        `json:"user_session_id" db:"user_session_id"`
	UserID        int        `json:"user_id" db:"user_id"`
	SessionToken  string     `json:"session_token" db:"session_token"`
	RefreshToken  *string    `json:"refresh_token,omitempty" db:"refresh_token"`
	ExpiresAt     time.Time  `json:"expires_at" db:"expires_at"`
	DeviceInfo    *string    `json:"device_info,omitempty" db:"device_info"`
	IPAddress     *string    `json:"ip_address,omitempty" db:"ip_address"`
	Location      *string    `json:"location,omitempty" db:"location"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserWithRoles représente un utilisateur avec ses rôles
type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}

// UserSessionWithUser représente une session avec les informations de l'utilisateur
type UserSessionWithUser struct {
	UserSession
	User User `json:"user"`
}

// Structures pour les requêtes
type CreateUserRequest struct {
	Lastname  string `json:"lastname" binding:"required"`
	Firstname string `json:"firstname" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	Lastname  *string `json:"lastname,omitempty"`
	Firstname *string `json:"firstname,omitempty"`
	Email     *string `json:"email,omitempty"`
	Password  *string `json:"password,omitempty"`
}

type CreateCalendarRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type UpdateCalendarRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description *string   `json:"description,omitempty"`
	Start       time.Time `json:"start" binding:"required"`
	Duration    int       `json:"duration" binding:"required,min=1"`
	CalendarID  int       `json:"calendar_id" binding:"required"`
	Canceled    *bool     `json:"canceled,omitempty"`
}

type UpdateEventRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Start       *time.Time `json:"start,omitempty"`
	Duration    *int       `json:"duration,omitempty" binding:"omitempty,min=1"`
	Canceled    *bool      `json:"canceled,omitempty"`
}

// Structures pour les requêtes de filtrage des événements
type ListEventsRequest struct {
	FilterType string `json:"filter_type" binding:"required,oneof=month week day"`
	Date       string `json:"date" binding:"required"` // Format: "2024-01-15" pour jour, "2024-01" pour mois, "2024-W01" pour semaine
}

// Structures pour l'authentification et les sessions
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User         User      `json:"user"`
	SessionToken string    `json:"session_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Roles        []Role    `json:"roles"`
}

type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type AssignRoleRequest struct {
	UserID int `json:"user_id" binding:"required"`
	RoleID int `json:"role_id" binding:"required"`
}

type RevokeRoleRequest struct {
	UserID int `json:"user_id" binding:"required"`
	RoleID int `json:"role_id" binding:"required"`
}

type CreateSessionRequest struct {
	UserID     int     `json:"user_id" binding:"required"`
	DeviceInfo *string `json:"device_info,omitempty"`
	IPAddress  *string `json:"ip_address,omitempty"`
	Location   *string `json:"location,omitempty"`
}

// StringPtr retourne un pointeur vers la chaîne passée en argument.
func StringPtr(s string) *string { return &s }

// IntPtr retourne un pointeur vers l'entier passé en argument.
func IntPtr(i int) *int { return &i }

// BoolPtr retourne un pointeur vers le booléen passé en argument.
func BoolPtr(b bool) *bool { return &b }
