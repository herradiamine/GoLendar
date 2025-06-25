package common

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetUserFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	user := User{UserID: 1, Email: "test@test.com"}
	c.Set("auth_user", user)
	res, ok := GetUserFromContext(c)
	if !ok || res.UserID != user.UserID {
		t.Errorf("GetUserFromContext() = %v, %v, want %v, true", res, ok, user)
	}
}

func TestGetCalendarFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	calendar := Calendar{CalendarID: 1, Title: "Test"}
	c.Set("calendar", calendar)
	res, ok := GetCalendarFromContext(c)
	if !ok || res.CalendarID != calendar.CalendarID {
		t.Errorf("GetCalendarFromContext() = %v, %v, want %v, true", res, ok, calendar)
	}
}

func TestGetEventFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	event := Event{EventID: 1, Title: "Test Event"}
	c.Set("event", event)
	res, ok := GetEventFromContext(c)
	if !ok || res.EventID != event.EventID {
		t.Errorf("GetEventFromContext() = %v, %v, want %v, true", res, ok, event)
	}
}
