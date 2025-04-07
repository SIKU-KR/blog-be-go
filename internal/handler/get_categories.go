package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"bumsiku/internal/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCategoriesResponse 카테고리 목록 응답 구조체
type GetCategoriesResponse struct {
	Categories []model.Category `json:"categories"` // 카테고리 목록
}

// @Summary     카테고리 목록 조회
// @Description 블로그에 등록된 모든 카테고리를 순서대로 조회합니다
// @Tags        카테고리
// @Accept      json
// @Produce     json
// @Success     200 {object} GetCategoriesResponse
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /categories [get]
func GetCategories(categoryRepo repository.CategoryRepositoryInterface, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 카테고리 목록 조회
		categories, err := categoryRepo.GetCategories(c.Request.Context())
		if err != nil {
			contextInfo := map[string]string{
				"handler":  "GetCategories",
				"step":     "카테고리 목록 조회",
				"clientIP": c.ClientIP(),
			}
			SendInternalServerErrorWithLogging(c, logger, "카테고리 목록 조회에 실패했습니다", err, contextInfo)
			return
		}

		response := GetCategoriesResponse{
			Categories: categories,
		}

		// 성공 로깅
		logger.Info(c.Request.Context(), "카테고리 목록 조회 성공", map[string]string{
			"handler":       "GetCategories",
			"categoryCount": fmt.Sprintf("%d", len(categories)),
			"clientIP":      c.ClientIP(),
		})

		SendSuccess(c, http.StatusOK, response)
	}
}
