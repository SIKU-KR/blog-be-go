package handler

import (
	"bumsiku/internal/repository"
	"bumsiku/internal/utils"
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
func GetPostByID(postRepo repository.PostRepositoryInterface, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		if postID == "" {
			contextInfo := map[string]string{
				"handler": "GetPostByID",
				"step":    "파라미터 검증",
			}
			SendBadRequestErrorWithLogging(c, logger, "게시글 ID가 필요합니다", nil, contextInfo)
			return
		}

		post, err := postRepo.GetPostByID(c.Request.Context(), postID)
		if err != nil {
			contextInfo := map[string]string{
				"handler": "GetPostByID",
				"step":    "게시글 조회",
				"postID":  postID,
			}
			SendInternalServerErrorWithLogging(c, logger, "게시글 조회에 실패했습니다", err, contextInfo)
			return
		}

		if post == nil {
			contextInfo := map[string]string{
				"handler": "GetPostByID",
				"step":    "결과 확인",
				"postID":  postID,
			}
			SendNotFoundErrorWithLogging(c, logger, "게시글을 찾을 수 없습니다", nil, contextInfo)
			return
		}

		// 성공 로깅 - 성능 모니터링에 유용
		logger.Info(c.Request.Context(), "게시글 상세 조회 성공", map[string]string{
			"handler":  "GetPostByID",
			"postID":   postID,
			"title":    post.Title,
			"category": post.Category,
			"clientIP": c.ClientIP(),
		})

		SendSuccess(c, http.StatusOK, post)
	}
}
