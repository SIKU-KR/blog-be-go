package handler

import (
	"bumsiku/internal/repository"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockGetPosts(repo *mockPostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 쿼리 파라미터 추출
		var category *string
		categoryParam := c.Query("category")
		if categoryParam != "" {
			category = &categoryParam
		}

		// 페이지네이션 파라미터
		page := int32(1)
		pageSize := int32(10)

		pageParam := c.Query("page")
		if pageParam != "" {
			if parsedPage, err := strconv.ParseInt(pageParam, 10, 32); err == nil && parsedPage > 0 {
				page = int32(parsedPage)
			}
		}

		pageSizeParam := c.Query("pageSize")
		if pageSizeParam != "" {
			if parsedPageSize, err := strconv.ParseInt(pageSizeParam, 10, 32); err == nil && parsedPageSize > 0 {
				pageSize = int32(parsedPageSize)
			}
		}

		// 저장소에서 게시글 조회
		input := &repository.GetPostsInput{
			Category: category,
			Page:     page,
			PageSize: pageSize,
		}

		output, err := repo.GetPosts(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "게시글 목록 조회에 실패했습니다",
				},
			})
			return
		}

		// 총 페이지 수 계산
		totalPages := output.TotalCount / int64(pageSize)
		if output.TotalCount%int64(pageSize) > 0 {
			totalPages++
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"posts":       output.Posts,
				"totalCount":  output.TotalCount,
				"currentPage": page,
				"totalPages":  totalPages,
			},
		})
	}
}

// [GIVEN] 정상적인 게시글 목록이 있는 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 게시글 목록 반환 확인
func TestGetPosts_Success(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts", "")

	MockGetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	posts := data["posts"].([]interface{})
	assert.Equal(t, 2, len(posts))
	assert.Equal(t, float64(2), data["totalCount"])
	assert.Equal(t, float64(1), data["currentPage"])
	assert.Equal(t, float64(1), data["totalPages"])
}

// [GIVEN] 카테고리 필터가 적용된 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 필터링된 게시글 목록 반환 확인
func TestGetPosts_WithCategory(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts?category=tech", "")

	MockGetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	posts := data["posts"].([]interface{})
	assert.Equal(t, 1, len(posts))
	assert.Equal(t, float64(1), data["totalCount"])
	assert.Equal(t, float64(1), data["currentPage"])
	assert.Equal(t, float64(1), data["totalPages"])
}

// [GIVEN] 페이지네이션이 적용된 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 페이지네이션이 적용된 응답 반환 확인
func TestGetPosts_WithPagination(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts?page=1&pageSize=1", "")

	MockGetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	posts := data["posts"].([]interface{})
	assert.Equal(t, 1, len(posts))
	assert.Equal(t, float64(2), data["totalCount"])
	assert.Equal(t, float64(1), data["currentPage"])
	assert.Equal(t, float64(2), data["totalPages"])
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetPosts_Error(t *testing.T) {
	// Given
	mockRepo := &mockPostRepository{err: errors.New("database error")}

	// When
	c, w := SetupTestContext("GET", "/posts", "")

	MockGetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Equal(t, "게시글 목록 조회에 실패했습니다", errorData["message"])
}
