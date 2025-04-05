package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateCategoryRequest는 카테고리 추가/수정 요청 구조체입니다.
type UpdateCategoryRequest struct {
	Category string `json:"category" binding:"required" example:"tech"` // 카테고리 식별자
	Order    int    `json:"order" binding:"required" example:"1"`       // 카테고리 정렬 순서
}

// @Summary     카테고리 추가/수정
// @Description 블로그 카테고리를 추가하거나 수정합니다 (관리자 전용)
// @Tags        카테고리
// @Accept      json
// @Produce     json
// @Security    AdminAuth
// @Param       request body UpdateCategoryRequest true "카테고리 정보"
// @Success     200 {object} model.Category
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     401 {object} ErrorResponse "인증 실패"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /admin/categories [put]
// UpdateCategory는 관리자 전용 카테고리 추가/수정 핸들러입니다.
func UpdateCategory(categoryRepo repository.CategoryRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 요청 바디 검증
		var req UpdateCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			SendBadRequestError(c, "요청 형식이 올바르지 않습니다: "+err.Error())
			return
		}

		// 2. 카테고리 모델 생성
		category := model.Category{
			Category:  req.Category,
			Order:     req.Order,
			CreatedAt: time.Time{}, // 레포지토리에서 신규 카테고리인 경우에만 설정됩니다
		}

		// 3. 카테고리 업데이트 또는 생성
		err := categoryRepo.UpsertCategory(c.Request.Context(), category)
		if err != nil {
			SendInternalServerError(c, "카테고리 업데이트에 실패했습니다: "+err.Error())
			return
		}

		// 4. 성공 응답
		SendSuccess(c, http.StatusOK, category)
	}
}
