package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPostById(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		postId := c.Param("postId")
		if postId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "게시글 ID가 필요합니다"})
			return
		}

		post, err := postRepo.GetPostById(c.Request.Context(), postId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "게시글 조회에 실패했습니다"})
			return
		}

		if post == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "게시글을 찾을 수 없습니다"})
			return
		}

		c.JSON(http.StatusOK, post)
	}
}
