package testutils

import (
	"time"
)

// Ce fichier contiendra les données de test réutilisables
// pour les tests du projet GoLendar

// ===== STRUCTS POUR LES TESTS DE SESSION =====

// SessionTestData contient les données de test pour les routes de session
type SessionTestData struct {
	ValidLoginRequest     map[string]interface{}
	InvalidLoginRequest   map[string]interface{}
	ValidRefreshRequest   map[string]interface{}
	InvalidRefreshRequest map[string]interface{}
	ValidLogoutHeaders    map[string]string
	InvalidLogoutHeaders  map[string]string
	ValidSessionHeaders   map[string]string
	InvalidSessionHeaders map[string]string
}

// ===== STRUCTS POUR LES TESTS D'UTILISATEUR =====

// UserTestData contient les données de test pour les routes d'utilisateur
type UserTestData struct {
	ValidCreateUserRequest    map[string]interface{}
	InvalidCreateUserRequest  map[string]interface{}
	ValidUpdateUserRequest    map[string]interface{}
	InvalidUpdateUserRequest  map[string]interface{}
	ValidAuthHeaders          map[string]string
	InvalidAuthHeaders        map[string]string
	ValidUserWithSpecialChars map[string]interface{}
	ValidUserWithComplexEmail map[string]interface{}
}

// ===== STRUCTS POUR LES TESTS DE RÔLES =====

// RoleTestData contient les données de test pour les routes de rôles
type RoleTestData struct {
	ValidCreateRoleRequest   map[string]interface{}
	InvalidCreateRoleRequest map[string]interface{}
	ValidUpdateRoleRequest   map[string]interface{}
	InvalidUpdateRoleRequest map[string]interface{}
	ValidAssignRoleRequest   map[string]interface{}
	InvalidAssignRoleRequest map[string]interface{}
	ValidRevokeRoleRequest   map[string]interface{}
	InvalidRevokeRoleRequest map[string]interface{}
	AdminHeaders             map[string]string
	NonAdminHeaders          map[string]string
}

// ===== STRUCTS POUR LES TESTS DE CALENDRIER =====

// CalendarTestData contient les données de test pour les routes de calendrier
type CalendarTestData struct {
	ValidCreateCalendarRequest   map[string]interface{}
	InvalidCreateCalendarRequest map[string]interface{}
	ValidUpdateCalendarRequest   map[string]interface{}
	InvalidUpdateCalendarRequest map[string]interface{}
	ValidCalendarHeaders         map[string]string
	InvalidCalendarHeaders       map[string]string
}

// ===== STRUCTS POUR LES TESTS D'ÉVÉNEMENTS =====

// EventTestData contient les données de test pour les routes d'événements
type EventTestData struct {
	ValidCreateEventRequest   map[string]interface{}
	InvalidCreateEventRequest map[string]interface{}
	ValidUpdateEventRequest   map[string]interface{}
	InvalidUpdateEventRequest map[string]interface{}
	ValidEventHeaders         map[string]string
	InvalidEventHeaders       map[string]string
	ValidEventFilters         map[string]interface{}
	InvalidEventFilters       map[string]interface{}
}

// ===== STRUCTS POUR LES TESTS DE LIAISONS USER-CALENDAR =====

// UserCalendarTestData contient les données de test pour les routes de liaisons utilisateur-calendrier
type UserCalendarTestData struct {
	ValidCreateUserCalendarRequest   map[string]interface{}
	InvalidCreateUserCalendarRequest map[string]interface{}
	ValidUpdateUserCalendarRequest   map[string]interface{}
	InvalidUpdateUserCalendarRequest map[string]interface{}
	AdminHeaders                     map[string]string
	NonAdminHeaders                  map[string]string
}

// ===== STRUCTS POUR LES TESTS DE MIDDLEWARE =====

// MiddlewareTestData contient les données de test pour les middlewares
type MiddlewareTestData struct {
	ValidAuthHeaders      map[string]string
	InvalidAuthHeaders    map[string]string
	ValidAdminHeaders     map[string]string
	NonAdminHeaders       map[string]string
	ExpiredTokenHeaders   map[string]string
	EmptyTokenHeaders     map[string]string
	MalformedTokenHeaders map[string]string
}

// ===== STRUCTS POUR LES TESTS GÉNÉRAUX =====

// CommonTestData contient les données de test communes à tous les modules
type CommonTestData struct {
	ValidJSONHeaders   map[string]string
	InvalidJSONHeaders map[string]string
	EmptyHeaders       map[string]string
	ValidContentType   string
	InvalidContentType string
	TestTime           time.Time
	TestDuration       time.Duration
}
