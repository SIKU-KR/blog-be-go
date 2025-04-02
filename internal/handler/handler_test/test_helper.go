package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gob.Register(time.Time{})
	gin.SetMode(gin.TestMode)
}

// MockPostRepository는 테스트에 사용되는 저장소 모의 객체입니다.
type mockPostRepository struct {
	posts     []model.Post
	nextToken *string
	err       error
}

func (m *mockPostRepository) GetPosts(ctx context.Context, input *repository.GetPostsInput) (*repository.GetPostsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	// 전체 게시글 수
	totalCount := int64(len(m.posts))
	
	// 카테고리 필터링
	filteredPosts := m.posts
	if input.Category != nil && *input.Category != "" {
		filtered := make([]model.Post, 0)
		for _, post := range m.posts {
			if post.Category == *input.Category {
				filtered = append(filtered, post)
			}
		}
		filteredPosts = filtered
		totalCount = int64(len(filteredPosts))
	}

	// 페이지네이션 적용
	start := (input.Page - 1) * input.PageSize
	end := start + input.PageSize
	if start >= int32(len(filteredPosts)) {
		return &repository.GetPostsOutput{
			Posts:      []model.Post{},
			TotalCount: totalCount,
		}, nil
	}
	if end > int32(len(filteredPosts)) {
		end = int32(len(filteredPosts))
	}

	return &repository.GetPostsOutput{
		Posts:      filteredPosts[start:end],
		TotalCount: totalCount,
	}, nil
}

func (m *mockPostRepository) GetPostByID(ctx context.Context, postID string) (*model.Post, error) {
	if m.err != nil {
		return nil, m.err
	}

	for _, post := range m.posts {
		if post.PostID == postID {
			return &post, nil
		}
	}

	return nil, nil
}

func (m *mockPostRepository) CreatePost(ctx context.Context, post *model.Post) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockPostRepository) UpdatePost(ctx context.Context, post *model.Post) error {
	if m.err != nil {
		return m.err
	}

	// 게시글 존재 여부 확인
	found := false
	for i, p := range m.posts {
		if p.PostID == post.PostID {
			found = true
			// 게시글 업데이트
			m.posts[i].Title = post.Title
			m.posts[i].Content = post.Content
			m.posts[i].Summary = post.Summary
			m.posts[i].Category = post.Category
			m.posts[i].UpdatedAt = post.UpdatedAt
			break
		}
	}

	if !found {
		return &repository.PostNotFoundError{PostID: post.PostID}
	}

	return nil
}

func (m *mockPostRepository) DeletePost(ctx context.Context, postID string) error {
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

// CreateTestPosts는 테스트용 게시글 데이터를 생성합니다.
func CreateTestPosts() []model.Post {
	now := time.Now()
	return []model.Post{
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

// CreateTestComments는 테스트용 댓글 데이터를 생성합니다.
func CreateTestComments() []model.Comment {
	now := time.Now()
	return []model.Comment{
		{
			CommentID: "comment1",
			PostID:    "post1",
			Content:   "첫 번째 댓글",
			Nickname:  "사용자1",
			CreatedAt: now,
		},
		{
			CommentID: "comment2",
			PostID:    "post1",
			Content:   "두 번째 댓글",
			Nickname:  "사용자2",
			CreatedAt: now.Add(time.Hour),
		},
		{
			CommentID: "comment3",
			PostID:    "post2",
			Content:   "다른 게시글의 댓글",
			Nickname:  "사용자3",
			CreatedAt: now.Add(2 * time.Hour),
		},
	}
}

// SetupTestContext는 테스트용 Gin 컨텍스트와 ResponseRecorder를 생성합니다.
func SetupTestContext(method, url, body string) (*gin.Context, *httptest.ResponseRecorder) {
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

// SetupTestContextWithSession은 세션이 있는 테스트용 Gin 컨텍스트와 ResponseRecorder를 생성합니다.
func SetupTestContextWithSession(method, url, body string) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := SetupTestContext(method, url, body)

	// 세션 미들웨어를 통해 세션 객체를 컨텍스트에 등록합니다.
	store := memstore.NewStore([]byte("secret"))
	sessionsMiddleware := sessions.Sessions("mysession", store)
	sessionsMiddleware(c)

	return c, w
}

// SetTestEnvironment는 테스트에 필요한 환경변수를 설정합니다.
func SetTestEnvironment() {
	os.Setenv("ADMIN_ID", "admin")
	os.Setenv("ADMIN_PW", "password")
}

// AssertResponseJSON은 응답 코드와 JSON 응답을 검증합니다.
func AssertResponseJSON(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedKey, expectedValue string) {
	assert.Equal(t, expectedStatus, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, resp[expectedKey])
}

// AssertJSONResponse는 응답 코드와 응답 바디를 검증합니다.
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedResponse interface{}) {
	assert.Equal(t, expectedStatus, w.Code)

	err := json.Unmarshal(w.Body.Bytes(), &expectedResponse)
	assert.NoError(t, err)
}

// CommentRepositoryMock은 repository.CommentRepositoryInterface를 구현하는 모의 객체입니다.
type CommentRepositoryMock struct {
	comments []model.Comment
	err      error
	// 생성된 댓글을 추적하기 위한 필드
	createdComment *model.Comment
}

func (m *CommentRepositoryMock) GetComments(ctx context.Context, input *repository.GetCommentsInput) ([]model.Comment, error) {
	if m.err != nil {
		return nil, m.err
	}

	// postID로 필터링
	if input != nil && input.PostID != nil {
		filteredComments := make([]model.Comment, 0)
		for _, comment := range m.comments {
			if comment.PostID == *input.PostID {
				filteredComments = append(filteredComments, comment)
			}
		}
		return filteredComments, nil
	}

	return m.comments, nil
}

func (m *CommentRepositoryMock) CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	if m.err != nil {
		return nil, m.err
	}

	// commentId 생성
	comment.CommentID = "new-comment-id"

	// 현재 시간 설정
	comment.CreatedAt = time.Now()

	// 생성된 댓글 저장 (테스트에서 검증 가능)
	m.createdComment = comment

	return comment, nil
}

func (m *CommentRepositoryMock) DeleteCommentsByPostID(ctx context.Context, postID string) error {
	if m.err != nil {
		return m.err
	}

	// 실제 삭제 로직은 테스트에서 중요하지 않으므로 성공만 반환
	return nil
}

func (m *CommentRepositoryMock) DeleteComment(ctx context.Context, commentID string) error {
	if m.err != nil {
		return m.err
	}

	// 실제 삭제 로직은 테스트에서 중요하지 않으므로 성공만 반환
	return nil
}
