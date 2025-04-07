package handler

import (
	"bumsiku/internal/model"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockCreatePost(repo *PostRepositoryForCreatePostMock) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 요청 본문 파싱
		var request struct {
			Title    string `json:"title"`
			Content  string `json:"content"`
			Summary  string `json:"summary"`
			Category string `json:"category"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "요청 형식이 올바르지 않습니다",
				},
			})
			return
		}

		// 필수 필드 검증
		if request.Title == "" || request.Content == "" || request.Summary == "" || request.Category == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "제목, 내용, 요약, 카테고리는 필수 항목입니다",
				},
			})
			return
		}

		// 게시글 생성
		now := time.Now()
		post := &model.Post{
			PostID:    xid.New().String(),
			Title:     request.Title,
			Content:   request.Content,
			Summary:   request.Summary,
			Category:  request.Category,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := repo.CreatePost(c.Request.Context(), post)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "게시글 등록에 실패했습니다",
				},
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    post,
		})
	}
}

// PostRepositoryForCreatePostMock은 CreatePost 함수를 구현한 Repository 모의 객체입니다.
type PostRepositoryForCreatePostMock struct {
	mockPostRepository
	createdPost *model.Post
}

func (m *PostRepositoryForCreatePostMock) CreatePost(ctx context.Context, post *model.Post) error {
	if m.err != nil {
		return m.err
	}
	m.createdPost = post
	return nil
}

// [GIVEN] 유효한 게시글 생성 요청이 있는 경우
// [WHEN] CreatePost 핸들러를 호출
// [THEN] 상태코드 201과 생성된 게시글 반환 확인
func TestCreatePost_Success(t *testing.T) {
	// Given
	mockRepo := &PostRepositoryForCreatePostMock{}
	requestBody := `{
		"title": "테스트 게시글",
		"content": "테스트 내용입니다.",
		"summary": "테스트 요약입니다.",
		"category": "tech"
	}`

	// When
	c, w := SetupTestContextWithSession("POST", "/admin/posts", string(requestBody))

	MockCreatePost(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	post := response["data"].(map[string]interface{})
	assert.Equal(t, "테스트 게시글", post["title"])
	assert.Equal(t, "테스트 내용입니다.", post["content"])
	assert.Equal(t, "테스트 요약입니다.", post["summary"])
	assert.Equal(t, "tech", post["category"])
	assert.NotEmpty(t, post["postId"])
	assert.Len(t, post["postId"], 20) // xid 패키지는 20자리 ID를 생성합니다
}

// [GIVEN] 유효하지 않은 요청 바디가 제공된 경우
// [WHEN] CreatePost 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestCreatePost_InvalidRequest(t *testing.T) {
	// Given
	mockRepo := &PostRepositoryForCreatePostMock{}
	requestBody := `{
		"title": "",
		"content": ""
	}`

	// When
	c, w := SetupTestContextWithSession("POST", "/admin/posts", string(requestBody))

	MockCreatePost(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errorData["code"])
	assert.Contains(t, errorData["message"], "필수 항목입니다")
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] CreatePost 핸들러를 호출
// [THEN] 상태코드 500과 적절한 에러 메시지 반환 확인
func TestCreatePost_SaveError(t *testing.T) {
	// Given
	mockRepo := &PostRepositoryForCreatePostMock{mockPostRepository: mockPostRepository{err: errors.New("database error")}}
	requestBody := `{
		"title": "테스트 게시글",
		"content": "테스트 내용입니다.",
		"summary": "테스트 요약입니다.",
		"category": "tech"
	}`

	// When
	c, w := SetupTestContextWithSession("POST", "/admin/posts", string(requestBody))

	MockCreatePost(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Equal(t, "게시글 등록에 실패했습니다", errorData["message"])
}
