package handler

import (
	"bumsiku/internal/handler"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// [GIVEN] 정상적인 댓글 목록이 있는 경우
// [WHEN] GetComments 핸들러를 호출
// [THEN] 상태코드 200과 댓글 목록 반환 확인
func TestGetComments_Success(t *testing.T) {
	// Given
	mockComments := CreateTestComments()
	mockRepo := &CommentRepositoryMock{comments: mockComments}

	// When
	c, w := SetupTestContext("GET", "/comments", "")
	handler.GetComments(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	comments := response["comments"].([]interface{})
	assert.Equal(t, 3, len(comments))
}

// [GIVEN] 특정 게시글의 댓글 필터링이 적용된 경우
// [WHEN] GetComments 핸들러를 호출(postId 쿼리 파라미터 있음)
// [THEN] 상태코드 200과 필터링된 댓글 목록 반환 확인
func TestGetComments_WithPostIdFilter(t *testing.T) {
	// Given
	mockComments := CreateTestComments()
	mockRepo := &CommentRepositoryMock{comments: mockComments}

	// When
	c, w := SetupTestContext("GET", "/comments?postId=post1", "")
	handler.GetComments(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	comments := response["comments"].([]interface{})
	assert.Equal(t, 2, len(comments)) // post1에 연결된 댓글은 2개
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetComments 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetComments_Error(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryMock{err: assert.AnError}

	// When
	c, w := SetupTestContext("GET", "/comments", "")
	handler.GetComments(mockRepo)(c)

	// Then
	AssertResponseJSON(t, w, http.StatusInternalServerError, "error", assert.AnError.Error())
}

// [GIVEN] 정상적인 특정 게시글의 댓글이 있는 경우
// [WHEN] GetCommentsByPostID 핸들러를 호출
// [THEN] 상태코드 200과 해당 게시글의 댓글 목록 반환 확인
func TestGetCommentsByPostID_Success(t *testing.T) {
	// Given
	mockComments := CreateTestComments()
	mockRepo := &CommentRepositoryMock{comments: mockComments}

	// When
	c, w := SetupTestContext("GET", "/posts/post1/comments", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	handler.GetCommentsByPostID(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	comments := response["comments"].([]interface{})
	assert.Equal(t, 2, len(comments)) // post1에 연결된 댓글은 2개
}

// [GIVEN] 게시글 ID가 비어있는 경우
// [WHEN] GetCommentsByPostID 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestGetCommentsByPostID_MissingId(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryMock{}

	// When
	c, w := SetupTestContext("GET", "/posts//comments", "")
	c.Params = []gin.Param{{Key: "id", Value: ""}}

	handler.GetCommentsByPostID(mockRepo)(c)

	// Then
	AssertResponseJSON(t, w, http.StatusBadRequest, "error", "게시글 ID가 필요합니다")
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetCommentsByPostID 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetCommentsByPostID_Error(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryMock{err: assert.AnError}

	// When
	c, w := SetupTestContext("GET", "/posts/post1/comments", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	handler.GetCommentsByPostID(mockRepo)(c)

	// Then
	AssertResponseJSON(t, w, http.StatusInternalServerError, "error", assert.AnError.Error())
}
