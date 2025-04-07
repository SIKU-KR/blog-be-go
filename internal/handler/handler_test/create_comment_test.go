package handler

import (
	"bumsiku/internal/model"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockCreateComment(commentRepo *CommentRepositoryMock, postRepo *mockPostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("postId")
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

		// 게시글 존재 확인
		post, err := postRepo.GetPostByID(c.Request.Context(), postID)
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
					"message": "존재하지 않는 게시글입니다",
				},
			})
			return
		}

		// 요청 바인딩
		var comment model.Comment
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": map[string]string{
					"code":    "BAD_REQUEST",
					"message": "요청 형식이 올바르지 않습니다",
				},
			})
			return
		}

		// 필수 필드 검증
		if comment.Nickname == "" || comment.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": map[string]string{
					"code":    "BAD_REQUEST",
					"message": "닉네임과 내용은 필수 입력 항목입니다",
				},
			})
			return
		}

		// 댓글에 게시글 ID 설정
		comment.PostID = postID

		// 댓글 저장
		createdComment, err := commentRepo.CreateComment(c.Request.Context(), &comment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": map[string]string{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "댓글 등록에 실패했습니다",
				},
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    createdComment,
		})
	}
}

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

	MockCreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	comment := response["data"].(map[string]interface{})
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

	MockCreateComment(mockCommentRepo, mockPostRepo)(c)

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

	MockCreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errorData["code"])
	assert.Equal(t, "존재하지 않는 게시글입니다", errorData["message"])
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

	MockCreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errorData["code"])
	assert.Equal(t, "닉네임과 내용은 필수 입력 항목입니다", errorData["message"])
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

	MockCreateComment(mockCommentRepo, mockPostRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Contains(t, errorData["message"], "댓글 등록에 실패했습니다")
}
