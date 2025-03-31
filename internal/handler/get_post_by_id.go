package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPostByID는 ID로 특정 게시물을 조회하는 핸들러 함수를 반환합니다.
func GetPostByID(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("postId")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "게시글 ID가 필요합니다"})
			return
		}

		post, err := postRepo.GetPostByID(c.Request.Context(), postID)
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
