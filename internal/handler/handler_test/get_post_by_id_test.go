package handler

import (
	"bumsiku/domain"
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
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts/nonexistent", "")
	c.Params = []gin.Param{{Key: "postId", Value: "nonexistent"}}

	handler.GetPostByID(mockRepo)(c)

	// Then
	AssertResponseJSON(t, w, http.StatusNotFound, "error", "게시글을 찾을 수 없습니다")
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
	AssertResponseJSON(t, w, http.StatusBadRequest, "error", "게시글 ID가 필요합니다")
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
	AssertResponseJSON(t, w, http.StatusInternalServerError, "error", "게시글 조회에 실패했습니다")
}
