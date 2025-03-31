package controller

import (
	"bumsiku/internal/container"
	"bumsiku/internal/handler"
	"bumsiku/internal/middleware"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const SESSION_STORE_NAME = "loginSession"

func SetupRouter(container *container.Container) *gin.Engine {
	router := gin.Default()
	router.Use(sessions.Sessions(SESSION_STORE_NAME, newSessionStore()))

	// Public Endpoints
	router.POST("/login", handler.PostLogin)
	router.GET("/posts", handler.GetPosts(container.PostRepository))
	router.GET("/posts/:id", handler.GetPostById(container.PostRepository))

	// Secured Endpoints
	admin := router.Group("/admin")
	admin.Use(middleware.SessionAuthMiddleware())

	return router
}

func newSessionStore() sessions.Store {
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		MaxAge:   7200,
		Path:     "/",
		HttpOnly: true,
	})
	return store
}
