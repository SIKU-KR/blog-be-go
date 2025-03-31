package handler

import (
	"bumsiku/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetPostsResponse struct {
	Posts     interface{} `json:"posts"`
	NextToken *string    `json:"nextToken,omitempty"`
}

func GetPosts(postRepo *repository.PostRepository) gin.HandlerFunc {
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