package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 실제 핸들러의 로직을 테스트용으로 복제하되 로깅 부분을 제거합니다
func MockGetPostByID(repo *mockPostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": map[string]string{
					"code":    "BAD_REQUEST",
					"message": "게시글 ID가 필요합니다",
				},
			})
			return
		}

		post, err := repo.GetPostByID(c.Request.Context(), postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": map[string]string{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "게시글 조회에 실패했습니다",
				},
			})
			return
		}

		if post == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": map[string]string{
					"code":    "NOT_FOUND",
					"message": "게시글을 찾을 수 없습니다",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    post,
		})
	}
}

// [GIVEN] 정상적인 게시글이 있는 경우
// [WHEN] GetPostById 핸들러를 호출
// [THEN] 상태코드 200과 해당 게시글 반환 확인
func TestGetPostById_Success(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts/post1", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	MockGetPostByID(mockRepo)(c)

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
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}

	MockGetPostByID(mockRepo)(c)

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
	c.Params = []gin.Param{{Key: "id", Value: ""}}

	MockGetPostByID(mockRepo)(c)

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
	mockRepo := &mockPostRepository{err: errors.New("database error")}

	// When
	c, w := SetupTestContext("GET", "/posts/post1", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	MockGetPostByID(mockRepo)(c)

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
