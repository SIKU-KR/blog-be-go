package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPostByID(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("postId")
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
