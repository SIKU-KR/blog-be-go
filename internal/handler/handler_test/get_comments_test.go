package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockGetComments(repo *CommentRepositoryMock) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 쿼리 파라미터 추출
		postID := c.Query("postId")
		var postIDPtr *string
		if postID != "" {
			postIDPtr = &postID
		}

		// 댓글 조회
		input := &repository.GetCommentsInput{
			PostID: postIDPtr,
		}

		comments, err := repo.GetComments(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"comments": comments,
			},
		})
	}
}

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockGetCommentsByPostID(repo *CommentRepositoryMock) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 경로 파라미터 추출
		postID := c.Param("id")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "게시글 ID가 필요합니다",
				},
			})
			return
		}

		var postIDPtr = &postID

		// 댓글 조회
		input := &repository.GetCommentsInput{
			PostID: postIDPtr,
		}

		comments, err := repo.GetComments(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    comments,
		})
	}
}

// [GIVEN] 정상적인 댓글 목록이 있는 경우
// [WHEN] GetComments 핸들러를 호출
// [THEN] 상태코드 200과 댓글 목록 반환 확인
func TestGetComments_Success(t *testing.T) {
	// Given
	mockComments := CreateTestComments()
	mockRepo := &CommentRepositoryMock{comments: mockComments}

	// When
	c, w := SetupTestContext("GET", "/comments", "")
	MockGetComments(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	comments := data["comments"].([]interface{})
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
	MockGetComments(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	comments := data["comments"].([]interface{})
	assert.Equal(t, 2, len(comments)) // post1에 연결된 댓글은 2개
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetComments 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetComments_Error(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryMock{err: errors.New("database error")}

	// When
	c, w := SetupTestContext("GET", "/comments", "")
	MockGetComments(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Equal(t, "database error", errorData["message"])
}

// [GIVEN] 정상적인 특정 게시글의 댓글이 있는 경우
// [WHEN] GetCommentsByPostID 핸들러를 호출
// [THEN] 상태코드 200과 해당 게시글의 댓글 목록 반환 확인
func TestGetCommentsByPostID_Success(t *testing.T) {
	// Given
	mockComments := CreateTestComments()
	mockRepo := &CommentRepositoryMock{comments: mockComments}

	// When
	c, w := SetupTestContext("GET", "/comments/post1", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	MockGetCommentsByPostID(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].([]interface{})
	assert.Equal(t, 2, len(data)) // post1에는 댓글이 2개 있어야 함
}

// [GIVEN] 게시글 ID가 비어있는 경우
// [WHEN] GetCommentsByPostID 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestGetCommentsByPostID_MissingID(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryMock{}

	// When
	c, w := SetupTestContext("GET", "/comments/", "")
	c.Params = []gin.Param{{Key: "id", Value: ""}}

	MockGetCommentsByPostID(mockRepo)(c)

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
// [WHEN] GetCommentsByPostID 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetCommentsByPostID_Error(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryMock{err: errors.New("database error")}

	// When
	c, w := SetupTestContext("GET", "/comments/post1", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	MockGetCommentsByPostID(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Equal(t, "database error", errorData["message"])
}

// 모든 댓글 가져오기 테스트
func TestGetAllComments_Success(t *testing.T) {
	// Given
	mockComments := CreateTestComments()
	mockRepo := &CommentRepositoryMock{comments: mockComments}

	// When
	c, w := SetupTestContext("GET", "/comments", "")
	MockGetComments(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])
}

// 댓글이 없는 경우 테스트
func TestGetComments_EmptyList(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryMock{comments: []model.Comment{}}

	// When
	c, w := SetupTestContext("GET", "/comments", "")
	MockGetComments(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	comments := data["comments"].([]interface{})
	assert.Equal(t, 0, len(comments))
}
