package handler

import (
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCategoriesResponse 카테고리 목록 응답 구조체
type GetCategoriesResponse struct {
	Categories interface{} `json:"categories" swaggertype:"array,object"` // 카테고리 목록
}

// @Summary     카테고리 목록 조회
// @Description 블로그에 등록된 모든 카테고리를 순서대로 조회합니다
// @Tags        카테고리
// @Accept      json
// @Produce     json
// @Success     200 {object} GetCategoriesResponse
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /categories [get]
func GetCategories(categoryRepo repository.CategoryRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 카테고리 목록 조회
		categories, err := categoryRepo.GetCategories(c.Request.Context())
		if err != nil {
			SendInternalServerError(c, "카테고리 목록 조회에 실패했습니다")
			return
		}

		response := GetCategoriesResponse{
			Categories: categories,
		}

		SendSuccess(c, http.StatusOK, response)
	}
}
