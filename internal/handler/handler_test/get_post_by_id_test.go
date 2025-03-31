package handler_test

import (
	"bumsiku/domain"
	"bumsiku/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// [GIVEN] 정상적인 게시글이 있는 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 200과 해당 게시글 반환 확인
func TestGetPostById_Success(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	mockPosts := createTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "postId", Value: "post1"}}
	req := httptest.NewRequest("GET", "/posts/post1", nil)
	c.Request = req

	handler.GetPostById(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var post domain.Post
	err := json.Unmarshal(w.Body.Bytes(), &post)
	assert.NoError(t, err)
	assert.Equal(t, "post1", post.PostID)
	assert.Equal(t, "첫 번째 게시글", post.Title)
}

// [GIVEN] 존재하지 않는 게시글 ID가 제공된 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 404와 적절한 에러 메시지 반환 확인
func TestGetPostById_NotFound(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	mockPosts := createTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "postId", Value: "nonexistent"}}
	req := httptest.NewRequest("GET", "/posts/nonexistent", nil)
	c.Request = req

	handler.GetPostById(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "게시글을 찾을 수 없습니다", response["error"])
}

// [GIVEN] postId 파라미터가 비어있는 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestGetPostById_MissingId(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	mockRepo := &mockPostRepository{}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "postId", Value: ""}}
	req := httptest.NewRequest("GET", "/posts/", nil)
	c.Request = req

	handler.GetPostById(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "게시글 ID가 필요합니다", response["error"])
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetPostById_Error(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	mockRepo := &mockPostRepository{err: assert.AnError}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "postId", Value: "post1"}}
	req := httptest.NewRequest("GET", "/posts/post1", nil)
	c.Request = req

	handler.GetPostById(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "게시글 조회에 실패했습니다", response["error"])
}
