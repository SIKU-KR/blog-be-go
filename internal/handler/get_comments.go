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
			SendInternalServerError(c, err.Error())
			return
		}

		SendSuccess(c, http.StatusOK, map[string]interface{}{
			"comments": comments,
		})
	}
}

// @Summary     게시물 댓글 조회
// @Description 특정 게시물에 작성된 댓글 목록을 조회합니다
// @Tags        댓글
// @Accept      json
// @Produce     json
// @Param       id path string true "게시물 ID"
// @Success     200 {object} map[string]interface{} "댓글 목록"
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /comments/{id} [get]
// GetCommentsByPostID는 특정 게시글의 댓글만 조회하는 핸들러입니다.
// URL 파라미터로 게시글 ID를 받습니다.
func GetCommentsByPostID(commentRepo repository.CommentRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// URL 파라미터에서 postId 추출
		postID := c.Param("id")
		if postID == "" {
			SendBadRequestError(c, "게시글 ID가 필요합니다")
			return
		}

		// 댓글 조회
		input := repository.GetCommentsInput{
			PostID: &postID,
		}
		comments, err := commentRepo.GetComments(c.Request.Context(), &input)
		if err != nil {
			SendInternalServerError(c, err.Error())
			return
		}

		SendSuccess(c, http.StatusOK, map[string]interface{}{
			"comments": comments,
		})
	}
}
