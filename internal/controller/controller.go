package controller

import (
	"bumsiku/internal/container"
	"bumsiku/internal/handler"
	"bumsiku/internal/middleware"
	"bumsiku/internal/utils"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const SessionStoreName = "loginSession"

func SetupRouter(container *container.Container) *gin.Engine {
	// 기본 gin 엔진 대신 새 엔진 생성 (기본 미들웨어 없이)
	router := gin.New()

	// 로거 생성
	logger := utils.NewLogger(container.CloudWatchClient)

	// 로깅과 복구 미들웨어 추가
	router.Use(middleware.RecoveryWithLogger(logger))
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.ErrorHandlingMiddleware(logger))
	router.Use(sessions.Sessions(SessionStoreName, newSessionStore()))

	// Static 파일 제공
	router.StaticFile("/robots.txt", "./static/robots.txt")

	// sitemap.xml 제공
	router.GET("/sitemap.xml", handler.GetSitemap(container.PostRepository, container.CategoryRepository))

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
	admin.POST("/images", handler.UploadImage(container.S3Client))

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
