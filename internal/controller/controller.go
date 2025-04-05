package controller

import (
	"bumsiku/internal/container"
	"bumsiku/internal/handler"
	"bumsiku/internal/middleware"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const SessionStoreName = "loginSession"

func SetupRouter(container *container.Container) *gin.Engine {
	router := gin.Default()
	router.Use(sessions.Sessions(SessionStoreName, newSessionStore()))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public Endpoints
	router.POST("/login", handler.PostLogin)
	router.GET("/posts", handler.GetPosts(container.PostRepository))
	router.GET("/posts/:id", handler.GetPostByID(container.PostRepository))
	router.GET("/comments/:id", handler.GetCommentsByPostID(container.CommentRepository))
	router.POST("/comments/:postId", handler.CreateComment(container.CommentRepository, container.PostRepository))
	router.GET("/categories", handler.GetCategories(container.CategoryRepository))

	// Secured Endpoints
	admin := router.Group("/admin")
	admin.Use(middleware.SessionAuthMiddleware())
	admin.POST("/posts", handler.CreatePost(container.PostRepository))
	admin.PUT("/posts/:id", handler.UpdatePost(container.PostRepository))
	admin.DELETE("/posts/:id", handler.DeletePost(container.PostRepository, container.CommentRepository))
	admin.DELETE("/comments/:commentId", handler.DeleteComment(container.CommentRepository))
	admin.PUT("/categories", handler.UpdateCategory(container.CategoryRepository))

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
