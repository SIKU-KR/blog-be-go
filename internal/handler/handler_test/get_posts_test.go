package handler

import (
	"bumsiku/internal/handler"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// [GIVEN] 정상적인 게시글 목록이 있는 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 게시글 목록 반환 확인
func TestGetPosts_Success(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts", "")
	handler.GetPosts(mockRepo)(c)

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
	handler.GetPosts(mockRepo)(c)

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
	handler.GetPosts(mockRepo)(c)

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
	mockRepo := &mockPostRepository{err: assert.AnError}

	// When
	c, w := SetupTestContext("GET", "/posts", "")
	handler.GetPosts(mockRepo)(c)

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
