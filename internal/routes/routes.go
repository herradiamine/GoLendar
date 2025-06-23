package routes

import (
	"go-averroes/internal/calendar"
	"go-averroes/internal/calendar_event"
	"go-averroes/internal/middleware"
	"go-averroes/internal/user"
	"go-averroes/internal/user_calendar"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// CRUD pour la gestion des utilisateurs
	router.GET(
		"/user/:user_id",
		middleware.UserExistsMiddleware("user_id"),
		func(c *gin.Context) { user.User.Get(c) },
	)
	router.POST(
		"/user",
		func(c *gin.Context) { user.User.Add(c) },
	)
	router.PUT(
		"/user/:user_id",
		middleware.UserExistsMiddleware("user_id"),
		func(c *gin.Context) { user.User.Update(c) },
	)
	router.DELETE(
		"/user/:user_id",
		middleware.UserExistsMiddleware("user_id"),
		func(c *gin.Context) { user.User.Delete(c) },
	)

	// CRUD pour la gestion des liaisons entre l'utilisateur et ses calendriers
	router.GET(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { user_calendar.UserCalendar.Get(c) },
	)
	router.POST(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { user_calendar.UserCalendar.Add(c) },
	)
	router.PUT(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { user_calendar.UserCalendar.Update(c) },
	)
	router.DELETE(
		"/user-calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { user_calendar.UserCalendar.Delete(c) },
	)

	// CRUD pour la gestion des calendriers
	router.GET(
		"/calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { calendar.Calendar.Get(c) },
	)
	router.POST(
		"/calendar/:user_id",
		middleware.UserExistsMiddleware("user_id"),
		func(c *gin.Context) { calendar.Calendar.Add(c) },
	)
	router.PUT(
		"/calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { calendar.Calendar.Update(c) },
	)
	router.DELETE(
		"/calendar/:user_id/:calendar_id",
		middleware.UserExistsMiddleware("user_id"),
		middleware.CalendarExistsMiddleware("calendar_id"),
		func(c *gin.Context) { calendar.Calendar.Delete(c) },
	)

	// CRUD pour la gestion des événements appartenant à leurs calendriers
	router.GET("/calendar-event/:id", func(c *gin.Context) { calendar_event.CalendarEvent.Get(c) })
	router.POST("/calendar-event", func(c *gin.Context) { calendar_event.CalendarEvent.Add(c) })
	router.PUT("/calendar-event/:id", func(c *gin.Context) { calendar_event.CalendarEvent.Update(c) })
	router.DELETE("/calendar-event/:id", func(c *gin.Context) { calendar_event.CalendarEvent.Delete(c) })
}
