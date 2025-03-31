package handler

import (
	"bumsiku/domain"
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

// CreateTestPosts는 테스트용 게시글 데이터를 생성합니다.
func CreateTestPosts() []domain.Post {
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
