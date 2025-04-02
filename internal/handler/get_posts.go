package handler

import (
	"bumsiku/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetPostsResponse struct {
	Posts       interface{} `json:"posts"`
	TotalCount  int64      `json:"totalCount"`
	CurrentPage int32      `json:"currentPage"`
	TotalPages  int32      `json:"totalPages"`
}

func GetPosts(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
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
			Category:  categoryPtr,
			Page:      page,
			PageSize:  pageSize,
		})

		if err != nil {
			SendInternalServerError(c, "게시글 목록 조회에 실패했습니다")
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

		SendSuccess(c, http.StatusOK, response)
	}
}
