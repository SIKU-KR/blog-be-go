package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary     댓글 삭제
// @Description 블로그 댓글을 삭제합니다 (관리자 전용)
// @Tags        댓글
// @Accept      json
// @Produce     json
// @Security    AdminAuth
// @Param       commentId path string true "댓글 ID"
// @Success     200 {object} map[string]string "삭제 성공 메시지"
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     401 {object} ErrorResponse "인증 실패"
// @Failure     404 {object} ErrorResponse "댓글을 찾을 수 없음"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /admin/comments/{commentId} [delete]
// DeleteComment는 관리자 전용 댓글 삭제 핸들러입니다.
func DeleteComment(commentRepo repository.CommentRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 경로 파라미터 확인
		commentID := c.Param("commentId")
		if commentID == "" {
			SendBadRequestError(c, "댓글 ID가 필요합니다")
			return
		}

		// 2. 댓글 삭제
		err := commentRepo.DeleteComment(c.Request.Context(), commentID)
		if err != nil {
			// CommentNotFoundError 확인
			if _, ok := err.(*repository.CommentNotFoundError); ok {
				SendNotFoundError(c, err.Error())
				return
			}

			SendInternalServerError(c, "댓글 삭제에 실패했습니다: "+err.Error())
			return
		}

		// 3. 성공 응답
		SendSuccess(c, http.StatusOK, map[string]string{
			"message": "댓글이 성공적으로 삭제되었습니다",
		})
	}
}
