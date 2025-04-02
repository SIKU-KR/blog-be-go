package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetComments는 전체 댓글 또는 특정 게시글의 댓글을 조회하는 핸들러입니다.
// 쿼리 파라미터로 postId를 받아 해당 게시글의 댓글만 필터링할 수 있습니다.
func GetComments(commentRepo repository.CommentRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input repository.GetCommentsInput

		// 쿼리 파라미터에서 postId 추출
		postID := c.Query("postId")
		if postID != "" {
			input.PostID = &postID
		}

		// 댓글 조회
		comments, err := commentRepo.GetComments(c.Request.Context(), &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"comments": comments,
		})
	}
}

// GetCommentsByPostID는 특정 게시글의 댓글만 조회하는 핸들러입니다.
// URL 파라미터로 게시글 ID를 받습니다.
func GetCommentsByPostID(commentRepo repository.CommentRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// URL 파라미터에서 postId 추출
		postID := c.Param("id")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "게시글 ID가 필요합니다",
			})
			return
		}

		// 댓글 조회
		input := repository.GetCommentsInput{
			PostID: &postID,
		}
		comments, err := commentRepo.GetComments(c.Request.Context(), &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"comments": comments,
		})
	}
}
