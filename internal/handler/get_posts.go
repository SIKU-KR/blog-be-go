package handler

import (
	"bumsiku/internal/repository"
	"bumsiku/internal/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPostsResponse 게시물 목록 응답 구조체
type GetPostsResponse struct {
	Posts       interface{} `json:"posts" swaggertype:"array,object"` // 게시물 목록
	TotalCount  int64       `json:"totalCount" example:"100"`         // 전체 게시물 수
	CurrentPage int32       `json:"currentPage" example:"1"`          // 현재 페이지
	TotalPages  int32       `json:"totalPages" example:"10"`          // 전체 페이지 수
}

// @Summary     게시물 목록 조회
// @Description 블로그 게시물 목록을 페이지네이션하여 조회합니다
// @Tags        게시물
// @Accept      json
// @Produce     json
// @Param       category query string false "카테고리 필터"
// @Param       page query int false "페이지 번호 (기본값: 1)"
// @Param       pageSize query int false "페이지 크기 (기본값: 10)"
// @Success     200 {object} GetPostsResponse
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /posts [get]
func GetPosts(postRepo repository.PostRepositoryInterface, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 쿼리 파라미터 파싱
		category := c.Query("category")
		pageStr := c.Query("page")
		pageSizeStr := c.Query("pageSize")

		// 기본값 설정
		page := int32(1)
		pageSize := int32(10)

		// 페이지 파라미터 처리
		if p, err := strconv.ParseInt(pageStr, 10, 32); err == nil && p > 0 {
			page = int32(p)
		}

		// 페이지 크기 파라미터 처리
		if size, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil && size > 0 {
			pageSize = int32(size)
		}

		// 카테고리 파라미터 처리
		var categoryPtr *string
		if category != "" {
			categoryPtr = &category
		}

		// 게시글 목록 조회
		result, err := postRepo.GetPosts(c.Request.Context(), &repository.GetPostsInput{
			Category: categoryPtr,
			Page:     page,
			PageSize: pageSize,
		})

		if err != nil {
			contextInfo := map[string]string{
				"handler":  "GetPosts",
				"step":     "게시글 목록 조회",
				"page":     fmt.Sprintf("%d", page),
				"pageSize": fmt.Sprintf("%d", pageSize),
				"clientIP": c.ClientIP(),
			}

			if categoryPtr != nil {
				contextInfo["category"] = *categoryPtr
			}

			SendInternalServerErrorWithLogging(c, logger, "게시글 목록 조회에 실패했습니다", err, contextInfo)
			return
		}

		// 전체 페이지 수 계산
		totalPages := (result.TotalCount + int64(pageSize) - 1) / int64(pageSize)

		response := GetPostsResponse{
			Posts:       result.Posts,
			TotalCount:  result.TotalCount,
			CurrentPage: page,
			TotalPages:  int32(totalPages),
		}

		// 성공 로깅
		contextInfo := map[string]string{
			"handler":    "GetPosts",
			"page":       fmt.Sprintf("%d", page),
			"pageSize":   fmt.Sprintf("%d", pageSize),
			"totalCount": fmt.Sprintf("%d", result.TotalCount),
			"clientIP":   c.ClientIP(),
		}

		if categoryPtr != nil {
			contextInfo["category"] = *categoryPtr
		}

		logger.Info(c.Request.Context(), "게시글 목록 조회 성공", contextInfo)

		SendSuccess(c, http.StatusOK, response)
	}
}
