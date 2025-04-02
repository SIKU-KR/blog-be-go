package handler

import (
	"bumsiku/internal/handler"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// [GIVEN] 정상적인 게시글이 있는 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 200과 해당 게시글 반환 확인
func TestGetPostById_Success(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts/post1", "")
	c.Params = []gin.Param{{Key: "postId", Value: "post1"}}

	handler.GetPostByID(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])
	
	post := response["data"].(map[string]interface{})
	assert.Equal(t, "post1", post["postId"])
	assert.Equal(t, "첫 번째 게시글", post["title"])
}

// [GIVEN] 존재하지 않는 게시글 ID가 제공된 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 404와 적절한 에러 메시지 반환 확인
func TestGetPostById_NotFound(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts/nonexistent", "")
	c.Params = []gin.Param{{Key: "postId", Value: "nonexistent"}}

	handler.GetPostByID(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	
	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errorData["code"])
	assert.Equal(t, "게시글을 찾을 수 없습니다", errorData["message"])
}

// [GIVEN] postId 파라미터가 비어있는 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestGetPostById_MissingId(t *testing.T) {
	// Given
	mockRepo := &mockPostRepository{}

	// When
	c, w := SetupTestContext("GET", "/posts/", "")
	c.Params = []gin.Param{{Key: "postId", Value: ""}}

	handler.GetPostByID(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	
	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errorData["code"])
	assert.Equal(t, "게시글 ID가 필요합니다", errorData["message"])
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetPostById_Error(t *testing.T) {
	// Given
	mockRepo := &mockPostRepository{err: assert.AnError}

	// When
	c, w := SetupTestContext("GET", "/posts/post1", "")
	c.Params = []gin.Param{{Key: "postId", Value: "post1"}}

	handler.GetPostByID(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	
	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Equal(t, "게시글 조회에 실패했습니다", errorData["message"])
}
