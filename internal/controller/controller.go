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

// SessionStoreName은 세션 스토어의 이름을 정의합니다.
const SessionStoreName = "loginSession"

// SetupRouter는 애플리케이션의 라우터를 설정하고 반환합니다.
func SetupRouter(container *container.Container) *gin.Engine {
	router := gin.Default()
	router.Use(sessions.Sessions(SessionStoreName, newSessionStore()))

	// Public Endpoints
	router.POST("/login", handler.PostLogin)
	router.GET("/posts", handler.GetPosts(container.PostRepository))
	router.GET("/posts/:id", handler.GetPostByID(container.PostRepository))

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
