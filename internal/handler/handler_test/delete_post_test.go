package handler

import (
	"bumsiku/internal/handler"
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// PostRepositoryForDeletePostMock은 DeletePost 함수를 구현한 Repository 모의 객체입니다.
type PostRepositoryForDeletePostMock struct {
	mockPostRepository
	deletedPostID string
}

func (m *PostRepositoryForDeletePostMock) DeletePost(ctx context.Context, postID string) error {
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

	m.deletedPostID = postID
	return nil
}

// CommentRepositoryForDeletePostMock은 DeleteCommentsByPostID 함수를 구현한 Repository 모의 객체입니다.
type CommentRepositoryForDeletePostMock struct {
	err                   error
	deletedCommentsPostID string
}

func (m *CommentRepositoryForDeletePostMock) GetComments(ctx context.Context, input *repository.GetCommentsInput) ([]model.Comment, error) {
	return nil, nil
}

func (m *CommentRepositoryForDeletePostMock) CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	return nil, nil
}

func (m *CommentRepositoryForDeletePostMock) DeleteCommentsByPostID(ctx context.Context, postID string) error {
	if m.err != nil {
		return m.err
	}
	m.deletedCommentsPostID = postID
	return nil
}

func (m *CommentRepositoryForDeletePostMock) DeleteComment(ctx context.Context, commentID string) error {
	if m.err != nil {
		return m.err
	}

	return nil
}

// [GIVEN] 유효한 게시글 ID가 있는 경우
// [WHEN] DeletePost 핸들러를 호출
// [THEN] 상태코드 200과 성공 메시지 반환 확인
func TestDeletePost_Success(t *testing.T) {
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

	mockPostRepo := &PostRepositoryForDeletePostMock{
		mockPostRepository: mockPostRepository{
			posts: posts,
		},
	}
	mockCommentRepo := &CommentRepositoryForDeletePostMock{}

	// When
	c, w := setupTestContext("DELETE", "/admin/posts/post1", "")
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "post1")
	handler.DeletePost(mockPostRepo, mockCommentRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "게시글이 성공적으로 삭제되었습니다", data["message"])
	assert.Equal(t, "post1", mockPostRepo.deletedPostID)
	assert.Equal(t, "post1", mockCommentRepo.deletedCommentsPostID)
}

// [GIVEN] 존재하지 않는 게시글 ID로 요청한 경우
// [WHEN] DeletePost 핸들러를 호출
// [THEN] 상태코드 404와 적절한 에러 메시지 반환 확인
func TestDeletePost_PostNotFound(t *testing.T) {
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

	mockPostRepo := &PostRepositoryForDeletePostMock{
		mockPostRepository: mockPostRepository{
			posts: posts,
		},
	}
	mockCommentRepo := &CommentRepositoryForDeletePostMock{}

	// When
	c, w := setupTestContext("DELETE", "/admin/posts/non-existent", "")
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "non-existent")
	handler.DeletePost(mockPostRepo, mockCommentRepo)(c)

	// Then
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errorData["code"])
	assert.Contains(t, errorData["message"], "게시글을 찾을 수 없음")
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] DeletePost 핸들러를 호출
// [THEN] 상태코드 500과 적절한 에러 메시지 반환 확인
func TestDeletePost_DeleteError(t *testing.T) {
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

	mockPostRepo := &PostRepositoryForDeletePostMock{
		mockPostRepository: mockPostRepository{
			posts: posts,
			err:   assert.AnError,
		},
	}
	mockCommentRepo := &CommentRepositoryForDeletePostMock{}

	// When
	c, w := setupTestContext("DELETE", "/admin/posts/post1", "")
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "post1")
	handler.DeletePost(mockPostRepo, mockCommentRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Contains(t, errorData["message"], "게시글 삭제에 실패했습니다")
}

// [GIVEN] 댓글 삭제 중 오류가 발생하더라도
// [WHEN] DeletePost 핸들러를 호출
// [THEN] 상태코드 200과 성공 메시지 반환 (게시글 삭제는 성공으로 간주)
func TestDeletePost_CommentDeleteError(t *testing.T) {
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

	mockPostRepo := &PostRepositoryForDeletePostMock{
		mockPostRepository: mockPostRepository{
			posts: posts,
		},
	}
	mockCommentRepo := &CommentRepositoryForDeletePostMock{
		err: assert.AnError,
	}

	// When
	c, w := setupTestContext("DELETE", "/admin/posts/post1", "")
	c.Set("admin", true) // 인증 상태 모의
	c.AddParam("id", "post1")
	handler.DeletePost(mockPostRepo, mockCommentRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "게시글이 성공적으로 삭제되었습니다", data["message"])
	assert.Equal(t, "post1", mockPostRepo.deletedPostID)
}
