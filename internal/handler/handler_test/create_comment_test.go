package handler

import (
	"bumsiku/internal/handler"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// [GIVEN] 유효한 댓글 생성 요청이 있는 경우
// [WHEN] CreateComment 핸들러를 호출
// [THEN] 상태코드 201과 생성된 댓글 반환 확인
func TestCreateComment_Success(t *testing.T) {
	// Given
	mockCommentRepo := &CommentRepositoryMock{}
	mockPostRepo := &mockPostRepository{posts: CreateTestPosts()}
	requestBody := `{
		"nickname": "테스터",
		"content": "테스트 댓글입니다."
	}`

	// When
	c, w := SetupTestContext("POST", "/comments/post1", requestBody)
	c.Params = []gin.Param{{Key: "postId", Value: "post1"}}

	handler.CreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	comment := response["comment"].(map[string]interface{})
	assert.Equal(t, "post1", comment["postId"])
	assert.Equal(t, "테스터", comment["nickname"])
	assert.Equal(t, "테스트 댓글입니다.", comment["content"])
}

// [GIVEN] 게시글 ID가 비어있는 경우
// [WHEN] CreateComment 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestCreateComment_MissingPostId(t *testing.T) {
	// Given
	mockCommentRepo := &CommentRepositoryMock{}
	mockPostRepo := &mockPostRepository{}
	requestBody := `{
		"nickname": "테스터",
		"content": "테스트 댓글입니다."
	}`

	// When
	c, w := SetupTestContext("POST", "/comments/", requestBody)
	c.Params = []gin.Param{{Key: "postId", Value: ""}}

	handler.CreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	AssertResponseJSON(t, w, http.StatusBadRequest, "error", "게시글 ID가 필요합니다")
}

// [GIVEN] 존재하지 않는 게시글 ID가 제공된 경우
// [WHEN] CreateComment 핸들러를 호출
// [THEN] 상태코드 404와 적절한 에러 메시지 반환 확인
func TestCreateComment_PostNotFound(t *testing.T) {
	// Given
	mockCommentRepo := &CommentRepositoryMock{}
	mockPostRepo := &mockPostRepository{posts: CreateTestPosts()}
	requestBody := `{
		"nickname": "테스터",
		"content": "테스트 댓글입니다."
	}`

	// When
	c, w := SetupTestContext("POST", "/comments/nonexistent", requestBody)
	c.Params = []gin.Param{{Key: "postId", Value: "nonexistent"}}

	handler.CreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	AssertResponseJSON(t, w, http.StatusNotFound, "error", "존재하지 않는 게시글입니다")
}

// [GIVEN] 유효하지 않은 요청 바디가 제공된 경우
// [WHEN] CreateComment 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestCreateComment_InvalidRequest(t *testing.T) {
	// Given
	mockCommentRepo := &CommentRepositoryMock{}
	mockPostRepo := &mockPostRepository{posts: CreateTestPosts()}
	requestBody := `{
		"nickname": "",
		"content": ""
	}`

	// When
	c, w := SetupTestContext("POST", "/comments/post1", requestBody)
	c.Params = []gin.Param{{Key: "postId", Value: "post1"}}

	handler.CreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "요청 형식이 올바르지 않습니다")
}

// [GIVEN] 댓글 저장 중 오류가 발생하는 경우
// [WHEN] CreateComment 핸들러를 호출
// [THEN] 상태코드 500과 적절한 에러 메시지 반환 확인
func TestCreateComment_SaveError(t *testing.T) {
	// Given
	mockCommentRepo := &CommentRepositoryMock{err: assert.AnError}
	mockPostRepo := &mockPostRepository{posts: CreateTestPosts()}
	requestBody := `{
		"nickname": "테스터",
		"content": "테스트 댓글입니다."
	}`

	// When
	c, w := SetupTestContext("POST", "/comments/post1", requestBody)
	c.Params = []gin.Param{{Key: "postId", Value: "post1"}}

	handler.CreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "댓글 등록에 실패했습니다")
}
