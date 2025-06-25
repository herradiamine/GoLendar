package routes

import (
	"go-averroes/internal/calendar"
	"go-averroes/internal/calendar_event"
	"go-averroes/internal/middleware"
	"go-averroes/internal/role"
	"go-averroes/internal/session"
	"go-averroes/internal/user"
	"go-averroes/internal/user_calendar"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// Health check endpoint (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "GoLendar API",
		})
	})

	// ===== ROUTES D'AUTHENTIFICATION (publiques) =====
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", func(c *gin.Context) { session.Session.Login(c) })
		authGroup.POST("/refresh", func(c *gin.Context) { session.Session.RefreshToken(c) })
	}

	// ===== ROUTES D'AUTHENTIFICATION (protégées) =====
	authProtectedGroup := router.Group("/auth")
	authProtectedGroup.Use(middleware.AuthMiddleware())
	{
		authProtectedGroup.POST("/logout", func(c *gin.Context) { session.Session.Logout(c) })
		authProtectedGroup.GET("/me", func(c *gin.Context) { user.User.GetUserWithRoles(c) })
		authProtectedGroup.GET("/sessions", func(c *gin.Context) { session.Session.GetUserSessions(c) })
		authProtectedGroup.DELETE("/sessions/:session_id", func(c *gin.Context) { session.Session.DeleteSession(c) })
	}

	// ===== ROUTES DE GESTION DES UTILISATEURS =====
	userGroup := router.Group("/user")
	{
		// Création d'utilisateur (public - inscription)
		userGroup.POST("", func(c *gin.Context) { user.User.Add(c) })

		// Routes protégées par authentification
		userProtectedGroup := userGroup.Group("")
		userProtectedGroup.Use(middleware.AuthMiddleware())
		{
			// L'utilisateur peut accéder à ses propres données
			userProtectedGroup.GET("/me", func(c *gin.Context) { user.User.Get(c) })
			userProtectedGroup.PUT("/me", func(c *gin.Context) { user.User.Update(c) })
			userProtectedGroup.DELETE("/me", func(c *gin.Context) { user.User.Delete(c) })
		}

		// Routes admin pour gérer tous les utilisateurs
		userAdminGroup := userGroup.Group("")
		userAdminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			userAdminGroup.GET("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.Get(c) })
			userAdminGroup.PUT("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.Update(c) })
			userAdminGroup.DELETE("/:user_id", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.Delete(c) })
			userAdminGroup.GET("/:user_id/with-roles", middleware.UserExistsMiddleware("user_id"), func(c *gin.Context) { user.User.GetUserWithRoles(c) })
		}
	}

	// ===== ROUTES DE GESTION DES RÔLES (admin uniquement) =====
	roleGroup := router.Group("/roles")
	roleGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		roleGroup.GET("", func(c *gin.Context) { role.Role.List(c) })
		roleGroup.GET("/:id", func(c *gin.Context) { role.Role.Get(c) })
		roleGroup.POST("", func(c *gin.Context) { role.Role.Add(c) })
		roleGroup.PUT("/:id", func(c *gin.Context) { role.Role.Update(c) })
		roleGroup.DELETE("/:id", func(c *gin.Context) { role.Role.Delete(c) })
		roleGroup.POST("/assign", func(c *gin.Context) { role.Role.AssignRole(c) })
		roleGroup.POST("/revoke", func(c *gin.Context) { role.Role.RevokeRole(c) })
		roleGroup.GET("/user/:user_id", func(c *gin.Context) { role.Role.GetUserRoles(c) })
	}

	// ===== ROUTES DE GESTION DES LIAISONS USER-CALENDAR (admin uniquement) =====
	userCalendarGroup := router.Group("/user-calendar")
	userCalendarGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		userCalendarGroup.GET("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Get(c) },
		)
		userCalendarGroup.GET("/:user_id",
			middleware.UserExistsMiddleware("user_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.List(c) },
		)
		userCalendarGroup.POST("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Add(c) },
		)
		userCalendarGroup.PUT("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Update(c) },
		)
		userCalendarGroup.DELETE("/:user_id/:calendar_id",
			middleware.UserExistsMiddleware("user_id"),
			middleware.CalendarExistsMiddleware("calendar_id"),
			func(c *gin.Context) { user_calendar.UserCalendar.Delete(c) },
		)
	}

	// ===== ROUTES DE GESTION DES CALENDRERS =====
	calendarGroup := router.Group("/calendar")
	calendarGroup.Use(middleware.AuthMiddleware())
	{
		// L'utilisateur peut créer des calendriers
		calendarGroup.POST("", func(c *gin.Context) { calendar.Calendar.Add(c) })

		// L'utilisateur peut accéder aux calendriers auxquels il a accès
		calendarGroup.GET("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar.Calendar.Get(c) },
		)
		calendarGroup.PUT("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar.Calendar.Update(c) },
		)
		calendarGroup.DELETE("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar.Calendar.Delete(c) },
		)
	}

	// ===== ROUTES DE GESTION DES ÉVÉNEMENTS =====
	calendarEventGroup := router.Group("/calendar-event")
	calendarEventGroup.Use(middleware.AuthMiddleware())
	{
		// Toutes les routes d'événements nécessitent l'accès au calendrier
		calendarEventGroup.GET("/:calendar_id/:event_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			middleware.EventExistsMiddleware("event_id"),
			func(c *gin.Context) { calendar_event.CalendarEvent.Get(c) },
		)
		calendarEventGroup.GET("/:calendar_id/month/:year/:month",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.ListByMonth(c) },
		)
		calendarEventGroup.GET("/:calendar_id/week/:year/:week",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.ListByWeek(c) },
		)
		calendarEventGroup.GET("/:calendar_id/day/:year/:month/:day",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.ListByDay(c) },
		)
		calendarEventGroup.POST("/:calendar_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			func(c *gin.Context) { calendar_event.CalendarEvent.Add(c) },
		)
		calendarEventGroup.PUT("/:calendar_id/:event_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			middleware.EventExistsMiddleware("event_id"),
			func(c *gin.Context) { calendar_event.CalendarEvent.Update(c) },
		)
		calendarEventGroup.DELETE("/:calendar_id/:event_id",
			middleware.CalendarExistsMiddleware("calendar_id"),
			middleware.UserCanAccessCalendarMiddleware(),
			middleware.EventExistsMiddleware("event_id"),
			func(c *gin.Context) { calendar_event.CalendarEvent.Delete(c) },
		)
	}
}
