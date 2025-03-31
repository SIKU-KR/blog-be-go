package handler_test

import (
	"bumsiku/domain"
	"bumsiku/internal/handler"
	"bumsiku/internal/repository"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock PostRepository
type mockPostRepository struct {
	posts     []domain.Post
	nextToken *string
	err       error
}

func (m *mockPostRepository) GetPosts(ctx context.Context, input *repository.GetPostsInput) (*repository.GetPostsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &repository.GetPostsOutput{
		Posts:     m.posts,
		NextToken: m.nextToken,
	}, nil
}

func (m *mockPostRepository) GetPostById(ctx context.Context, postId string) (*domain.Post, error) {
	if m.err != nil {
		return nil, m.err
	}

	for _, post := range m.posts {
		if post.PostID == postId {
			return &post, nil
		}
	}

	return nil, nil
}

// 테스트용 데이터 생성
func createTestPosts() []domain.Post {
	now := time.Now()
	return []domain.Post{
		{
			PostID:    "post1",
			Title:     "첫 번째 게시글",
			Content:   "내용 1",
			Category:  "tech",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			PostID:    "post2",
			Title:     "두 번째 게시글",
			Content:   "내용 2",
			Category:  "life",
			CreatedAt: now.Add(time.Hour),
			UpdatedAt: now.Add(time.Hour),
		},
	}
}

// [GIVEN] 정상적인 게시글 목록이 있는 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 게시글 목록 반환 확인
func TestGetPosts_Success(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	mockPosts := createTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/posts", nil)
	c.Request = req

	handler.GetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response handler.GetPostsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	posts := response.Posts.([]interface{})
	assert.Equal(t, 2, len(posts))
}

// [GIVEN] 카테고리 필터가 적용된 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 필터링된 게시글 목록 반환 확인
func TestGetPosts_WithCategory(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	mockPosts := []domain.Post{createTestPosts()[0]} // tech 카테고리만
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/posts?category=tech", nil)
	c.Request = req

	handler.GetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response handler.GetPostsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	posts := response.Posts.([]interface{})
	assert.Equal(t, 1, len(posts))
}

// [GIVEN] 페이지네이션이 적용된 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 nextToken이 포함된 응답 반환 확인
func TestGetPosts_WithPagination(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	nextToken := "next_page_token"
	mockRepo := &mockPostRepository{
		posts:     createTestPosts()[:1],
		nextToken: &nextToken,
	}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/posts?pageSize=1", nil)
	c.Request = req

	handler.GetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response handler.GetPostsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, nextToken, *response.NextToken)
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetPosts_Error(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	mockRepo := &mockPostRepository{err: assert.AnError}

	// When
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/posts", nil)
	c.Request = req

	handler.GetPosts(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "게시글 목록 조회에 실패했습니다", response["error"])
}
