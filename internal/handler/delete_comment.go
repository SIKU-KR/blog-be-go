package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteComment는 관리자 전용 댓글 삭제 핸들러입니다.
func DeleteComment(commentRepo repository.CommentRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 경로 파라미터 확인
		commentID := c.Param("commentId")
		if commentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "댓글 ID가 필요합니다",
			})
			return
		}

		// 2. 댓글 삭제
		err := commentRepo.DeleteComment(c.Request.Context(), commentID)
		if err != nil {
			// CommentNotFoundError 확인
			if _, ok := err.(*repository.CommentNotFoundError); ok {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "댓글 삭제에 실패했습니다: " + err.Error(),
			})
			return
		}

		// 3. 성공 응답
		c.JSON(http.StatusOK, gin.H{
			"message": "댓글이 성공적으로 삭제되었습니다",
		})
	}
}
