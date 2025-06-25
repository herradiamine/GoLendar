package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserFromContext récupère l'utilisateur du contexte Gin.
// En cas d'échec, il envoie une réponse d'erreur et retourne false.
func GetUserFromContext(c *gin.Context) (User, bool) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, JSONResponse{
			Success: false,
			Error:   ErrInternalUserNotInContext,
		})
		return User{}, false
	}
	userData, ok := user.(User)
	if !ok {
		// This case should ideally not happen if middleware is set correctly
		c.JSON(http.StatusInternalServerError, JSONResponse{
			Success: false,
			Error:   ErrContextUserType,
		})
		return User{}, false
	}
	return userData, true
}

// GetCalendarFromContext récupère le calendrier du contexte Gin.
// En cas d'échec, il envoie une réponse d'erreur et retourne false.
func GetCalendarFromContext(c *gin.Context) (Calendar, bool) {
	calendar, exists := c.Get("calendar")
	if !exists {
		c.JSON(http.StatusInternalServerError, JSONResponse{
			Success: false,
			Error:   ErrInternalCalendarNotInContext,
		})
		return Calendar{}, false
	}
	calendarData, ok := calendar.(Calendar)
	if !ok {
		// This case should ideally not happen if middleware is set correctly
		c.JSON(http.StatusInternalServerError, JSONResponse{
			Success: false,
			Error:   ErrContextCalendarType,
		})
		return Calendar{}, false
	}
	return calendarData, true
}

// GetEventFromContext récupère l'événement du contexte Gin.
// En cas d'échec, il envoie une réponse d'erreur et retourne false.
func GetEventFromContext(c *gin.Context) (Event, bool) {
	event, exists := c.Get("event")
	if !exists {
		c.JSON(http.StatusInternalServerError, JSONResponse{
			Success: false,
			Error:   ErrEventRetrieval,
		})
		return Event{}, false
	}
	eventData, ok := event.(Event)
	if !ok {
		// This case should ideally not happen if middleware is set correctly
		c.JSON(http.StatusInternalServerError, JSONResponse{
			Success: false,
			Error:   ErrEventRetrieval,
		})
		return Event{}, false
	}
	return eventData, true
}
