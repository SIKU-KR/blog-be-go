package handler

import (
	"bumsiku/internal/repository"
	"bumsiku/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary     게시물 삭제
// @Description 블로그 게시물과 관련 댓글을 삭제합니다 (관리자 전용)
// @Tags        게시물
// @Accept      json
// @Produce     json
// @Security    AdminAuth
// @Param       id path string true "게시물 ID"
// @Success     200 {object} map[string]string "삭제 성공 메시지"
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     401 {object} ErrorResponse "인증 실패"
// @Failure     404 {object} ErrorResponse "게시물을 찾을 수 없음"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /admin/posts/{id} [delete]
// DeletePost는 관리자 전용 게시글 삭제 핸들러입니다.
func DeletePost(postRepo repository.PostRepositoryInterface, commentRepo repository.CommentRepositoryInterface, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 경로 파라미터 확인
		postID := c.Param("id")
		if postID == "" {
			contextInfo := map[string]string{
				"handler": "DeletePost",
				"step":    "경로 파라미터 확인",
			}
			SendBadRequestErrorWithLogging(c, logger, "게시글 ID가 필요합니다", nil, contextInfo)
			return
		}

		// 게시글 정보 먼저 조회 (로깅용)
		post, err := postRepo.GetPostByID(c.Request.Context(), postID)
		var title, category string
		if err == nil && post != nil {
			title = post.Title
			category = post.Category
		}

		// 2. 게시글 삭제
		err = postRepo.DeletePost(c.Request.Context(), postID)
		if err != nil {
			contextInfo := map[string]string{
				"handler": "DeletePost",
				"step":    "게시글 삭제",
				"postID":  postID,
			}
			if title != "" {
				contextInfo["title"] = title
			}
			if category != "" {
				contextInfo["category"] = category
			}

			// PostNotFoundError 확인
			if _, ok := err.(*repository.PostNotFoundError); ok {
				SendNotFoundErrorWithLogging(c, logger, "게시글을 찾을 수 없습니다", err, contextInfo)
				return
			}

			SendInternalServerErrorWithLogging(c, logger, "게시글 삭제에 실패했습니다", err, contextInfo)
			return
		}

		// 3. 연관된 댓글 삭제
		err = commentRepo.DeleteCommentsByPostID(c.Request.Context(), postID)
		if err != nil {
			// 댓글 삭제 실패 로그를 남기지만, 사용자에게는 게시글 삭제 성공으로 응답
			logger.Warn(c.Request.Context(), "게시글 삭제 성공 후 관련 댓글 삭제 실패", map[string]string{
				"handler": "DeletePost",
				"step":    "댓글 삭제",
				"postID":  postID,
				"error":   err.Error(),
			})
		}

		// 로그 남기기 - 성공 케이스
		contextInfo := map[string]string{
			"handler": "DeletePost",
			"postID":  postID,
		}
		if title != "" {
			contextInfo["title"] = title
		}
		if category != "" {
			contextInfo["category"] = category
		}
		logger.Info(c.Request.Context(), "게시글이 성공적으로 삭제되었습니다", contextInfo)

		// 4. 성공 응답
		SendSuccess(c, http.StatusOK, map[string]string{
			"message": "게시글이 성공적으로 삭제되었습니다",
		})
	}
}
