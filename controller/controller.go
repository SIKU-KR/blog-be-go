package controller

import (
	"os"

	"bumsiku/internal/handler"
	"bumsiku/internal/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const SESSION_STORE_NAME = "loginSession"

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(sessions.Sessions(SESSION_STORE_NAME, createSessionStore()))

	// Public Endpoints
	router.POST("/login", handler.PostLogin)

	// Secured Endpoints
	admin := router.Group("/admin")
	admin.Use(middleware.SessionAuthMiddleware())

	return router
}

func createSessionStore() sessions.Store {
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		MaxAge:   7200,
		Path:     "/",
		HttpOnly: true,
	})
	return store
}
