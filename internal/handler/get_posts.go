package handler

import (
	"bumsiku/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPostsResponse는 게시글 목록 조회 응답 데이터를 담는 구조체입니다.
type GetPostsResponse struct {
	Posts     interface{} `json:"posts"`
	NextToken *string     `json:"nextToken,omitempty"`
}

// GetPosts는 게시글 목록을 조회하는 핸들러 함수를 반환합니다.
func GetPosts(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 쿼리 파라미터 파싱
		category := c.Query("category")
		nextToken := c.Query("nextToken")
		pageSizeStr := c.Query("pageSize")

		var pageSize *int32
		if pageSizeStr != "" {
			if size, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil {
				pageSizeInt32 := int32(size)
				pageSize = &pageSizeInt32
			}
		}

		// 카테고리 파라미터 처리
		var categoryPtr *string
		if category != "" {
			categoryPtr = &category
		}

		// 다음 페이지 토큰 처리
		var nextTokenPtr *string
		if nextToken != "" {
			nextTokenPtr = &nextToken
		}

		// 게시글 목록 조회
		result, err := postRepo.GetPosts(c.Request.Context(), &repository.GetPostsInput{
			Category:  categoryPtr,
			NextToken: nextTokenPtr,
			PageSize:  pageSize,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "게시글 목록 조회에 실패했습니다"})
			return
		}

		c.JSON(http.StatusOK, GetPostsResponse{
			Posts:     result.Posts,
			NextToken: result.NextToken,
		})
	}
}
