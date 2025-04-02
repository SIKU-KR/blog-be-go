package handler

import (
	"bumsiku/internal/handler"
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 테스트용 헬퍼 함수
func setupTestContext(method, url, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	c.Request = req
	return c, w
}

// PostRepositoryForUpdatePostMock은 UpdatePost 함수를 위한 Repository 모의 객체입니다.
type PostRepositoryForUpdatePostMock struct {
	posts       []model.Post
	err         error
	updatedPost *model.Post
}

func (m *PostRepositoryForUpdatePostMock) GetPosts(ctx context.Context, input *repository.GetPostsInput) (*repository.GetPostsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &repository.GetPostsOutput{
		Posts:     m.posts,
		NextToken: nil,
	}, nil
}

func (m *PostRepositoryForUpdatePostMock) GetPostByID(ctx context.Context, postID string) (*model.Post, error) {
	if m.err != nil {
		return nil, m.err
	}

	// 업데이트된 게시글이 있으면 반환
	if m.updatedPost != nil && m.updatedPost.PostID == postID {
		now := time.Now()
		return &model.Post{
			PostID:    m.updatedPost.PostID,
			Title:     m.updatedPost.Title,
			Content:   m.updatedPost.Content,
			Summary:   m.updatedPost.Summary,
			Category:  m.updatedPost.Category,
			CreatedAt: now.Add(-time.Hour), // 1시간 전 생성
			UpdatedAt: now,                 // 지금 업데이트
		}, nil
	}

	// 기존 게시글 찾기
	for _, post := range m.posts {
		if post.PostID == postID {
			return &post, nil
		}
	}

	return nil, nil
}

func (m *PostRepositoryForUpdatePostMock) CreatePost(ctx context.Context, post *model.Post) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *PostRepositoryForUpdatePostMock) UpdatePost(ctx context.Context, post *model.Post) error {
	if m.err != nil {
		return m.err
	}

	// 게시글 존재 여부 확인
	found := false
	for _, p := range m.posts {
		if p.PostID == post.PostID {
			found = true
			break
		}
	}

	if !found {
		return &repository.PostNotFoundError{PostID: post.PostID}
	}

	m.updatedPost = post
	return nil
}

func (m *PostRepositoryForUpdatePostMock) DeletePost(ctx context.Context, postID string) error {
	if m.err != nil {
		return m.err
	}

	// 게시글 존재 여부 확인
	found := false
	for _, p := range m.posts {
		if p.PostID == postID {
			found = true
			break
		}
	}

	if !found {
		return &repository.PostNotFoundError{PostID: postID}
	}

	return nil
}

// [GIVEN] 유효한 게시글 수정 요청이 있는 경우
// [WHEN] UpdatePost 핸들러를 호출
// [THEN] 상태코드 200과 수정된 게시글 반환 확인
func TestUpdatePost_Success(t *testing.T) {
	// Given
	now := time.Now()
	posts := []model.Post{
		{
			PostID:    "post1",
			Title:     "첫 번째 게시글",
			Content:   "내용 1",
			Summary:   "요약 1",
			Category:  "tech",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			PostID:    "post2",
			Title:     "두 번째 게시글",
			Content:   "내용 2",
			Summary:   "요약 2",
			Category:  "life",
			CreatedAt: now.Add(time.Hour),
			UpdatedAt: now.Add(time.Hour),
		},
	}

	mockRepo := &PostRepositoryForUpdatePostMock{
		posts: posts,
	}

	requestBody := `{
		"title": "수정된 게시글",
		"content": "수정된 내용입니다.",
		"summary": "수정된 요약입니다.",
		"category": "life"
	}`

	// When
	c, w := setupTestContext("PUT", "/admin/posts/post1", requestBody)
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "post1")
	handler.UpdatePost(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	post := response["post"].(map[string]interface{})
	assert.Equal(t, "post1", post["postId"])
	assert.Equal(t, "수정된 게시글", post["title"])
	assert.Equal(t, "수정된 내용입니다.", post["content"])
	assert.Equal(t, "수정된 요약입니다.", post["summary"])
	assert.Equal(t, "life", post["category"])
}

// [GIVEN] 존재하지 않는 게시글 ID로 요청한 경우
// [WHEN] UpdatePost 핸들러를 호출
// [THEN] 상태코드 404와 적절한 에러 메시지 반환 확인
func TestUpdatePost_PostNotFound(t *testing.T) {
	// Given
	now := time.Now()
	posts := []model.Post{
		{
			PostID:    "post1",
			Title:     "첫 번째 게시글",
			Content:   "내용 1",
			Summary:   "요약 1",
			Category:  "tech",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockRepo := &PostRepositoryForUpdatePostMock{
		posts: posts,
	}

	requestBody := `{
		"title": "수정된 게시글",
		"content": "수정된 내용입니다.",
		"summary": "수정된 요약입니다.",
		"category": "life"
	}`

	// When
	c, w := setupTestContext("PUT", "/admin/posts/non-existent", requestBody)
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "non-existent")
	handler.UpdatePost(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "게시글을 찾을 수 없음")
}

// [GIVEN] 요청 형식이 올바르지 않은 경우
// [WHEN] UpdatePost 핸들러를 호출
// [THEN] 상태코드 400과 적절한 에러 메시지 반환 확인
func TestUpdatePost_InvalidRequest(t *testing.T) {
	// Given
	now := time.Now()
	posts := []model.Post{
		{
			PostID:    "post1",
			Title:     "첫 번째 게시글",
			Content:   "내용 1",
			Summary:   "요약 1",
			Category:  "tech",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockRepo := &PostRepositoryForUpdatePostMock{
		posts: posts,
	}

	requestBody := `{
		"title": "", 
		"content": ""
	}`

	// When
	c, w := setupTestContext("PUT", "/admin/posts/post1", requestBody)
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "post1")
	handler.UpdatePost(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "요청 형식이 올바르지 않습니다")
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] UpdatePost 핸들러를 호출
// [THEN] 상태코드 500과 적절한 에러 메시지 반환 확인
func TestUpdatePost_UpdateError(t *testing.T) {
	// Given
	now := time.Now()
	posts := []model.Post{
		{
			PostID:    "post1",
			Title:     "첫 번째 게시글",
			Content:   "내용 1",
			Summary:   "요약 1",
			Category:  "tech",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockRepo := &PostRepositoryForUpdatePostMock{
		posts: posts,
		err:   assert.AnError,
	}

	requestBody := `{
		"title": "수정된 게시글",
		"content": "수정된 내용입니다.",
		"summary": "수정된 요약입니다.",
		"category": "life"
	}`

	// When
	c, w := setupTestContext("PUT", "/admin/posts/post1", requestBody)
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "post1")
	handler.UpdatePost(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "게시글 수정에 실패했습니다")
}
