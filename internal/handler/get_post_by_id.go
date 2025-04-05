package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary     게시물 상세 조회
// @Description 특정 ID의 게시물 상세 정보를 조회합니다
// @Tags        게시물
// @Accept      json
// @Produce     json
// @Param       id path string true "게시물 ID"
// @Success     200 {object} model.Post
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     404 {object} ErrorResponse "게시물을 찾을 수 없음"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /posts/{id} [get]
func GetPostByID(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		if postID == "" {
			SendBadRequestError(c, "게시글 ID가 필요합니다")
			return
		}

		post, err := postRepo.GetPostByID(c.Request.Context(), postID)
		if err != nil {
			SendInternalServerError(c, "게시글 조회에 실패했습니다")
			return
		}

		if post == nil {
			SendNotFoundError(c, "게시글을 찾을 수 없습니다")
			return
		}

		SendSuccess(c, http.StatusOK, post)
	}
}
