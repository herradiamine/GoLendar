package routes

import (
	"github.com/gin-gonic/gin"
	"go-averroes/internal/calendar"
	"go-averroes/internal/calendar_event"
	"go-averroes/internal/user"
	"go-averroes/internal/user_calendar"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/user/:id", func(c *gin.Context) { user.User.Get(c) })
	router.POST("/user", func(c *gin.Context) { user.User.Add(c) })
	router.PUT("/user/:id", func(c *gin.Context) { user.User.Update(c) })
	router.DELETE("/user/:id", func(c *gin.Context) { user.User.Delete(c) })

	router.GET("/user-calendar/:id", func(c *gin.Context) { user_calendar.UserCalendar.Get(c) })
	router.POST("/user-calendar", func(c *gin.Context) { user_calendar.UserCalendar.Add(c) })
	router.PUT("/user-calendar/:id", func(c *gin.Context) { user_calendar.UserCalendar.Update(c) })
	router.DELETE("/user-calendar/:id", func(c *gin.Context) { user_calendar.UserCalendar.Delete(c) })

	router.GET("/calendar/:id", func(c *gin.Context) { calendar.Calendar.Get(c) })
	router.POST("/calendar", func(c *gin.Context) { calendar.Calendar.Add(c) })
	router.PUT("/calendar/:id", func(c *gin.Context) { calendar.Calendar.Update(c) })
	router.DELETE("/calendar/:id", func(c *gin.Context) { calendar.Calendar.Delete(c) })

	router.GET("/calendar-event/:id", func(c *gin.Context) { calendar_event.CalendarEvent.Get(c) })
	router.POST("/calendar-event", func(c *gin.Context) { calendar_event.CalendarEvent.Add(c) })
	router.PUT("/calendar-event/:id", func(c *gin.Context) { calendar_event.CalendarEvent.Update(c) })
	router.DELETE("/calendar-event/:id", func(c *gin.Context) { calendar_event.CalendarEvent.Delete(c) })
}
