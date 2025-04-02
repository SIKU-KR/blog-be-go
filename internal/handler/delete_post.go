package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeletePost는 관리자 전용 게시글 삭제 핸들러입니다.
func DeletePost(postRepo repository.PostRepositoryInterface, commentRepo repository.CommentRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 경로 파라미터 확인
		postID := c.Param("id")
		if postID == "" {
			SendBadRequestError(c, "게시글 ID가 필요합니다")
			return
		}

		// 2. 게시글 삭제
		err := postRepo.DeletePost(c.Request.Context(), postID)
		if err != nil {
			// PostNotFoundError 확인
			if _, ok := err.(*repository.PostNotFoundError); ok {
				SendNotFoundError(c, err.Error())
				return
			}

			SendInternalServerError(c, "게시글 삭제에 실패했습니다: "+err.Error())
			return
		}

		// 3. 연관된 댓글 삭제
		err = commentRepo.DeleteCommentsByPostID(c.Request.Context(), postID)
		if err != nil {
			// 댓글 삭제 실패 로그를 남길 수 있지만, 사용자에게는 게시글 삭제 성공으로 응답
			// 로깅 구현은 생략
		}

		// 4. 성공 응답
		SendSuccess(c, http.StatusOK, map[string]string{
			"message": "게시글이 성공적으로 삭제되었습니다",
		})
	}
}
