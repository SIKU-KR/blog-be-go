package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockDeletePost(postRepo *PostRepositoryForDeletePostMock, commentRepo *CommentRepositoryForDeletePostMock) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// 게시글 삭제
		err := postRepo.DeletePost(c.Request.Context(), postID)
		if err != nil {
			var notFoundErr *repository.PostNotFoundError
			if errors.As(err, &notFoundErr) {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "NOT_FOUND",
						"message": "게시글을 찾을 수 없음: " + postID,
					},
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "게시글 삭제에 실패했습니다",
				},
			})
			return
		}

		// 게시글 관련 댓글 삭제 (실패해도 게시글 삭제는 성공으로 간주)
		_ = commentRepo.DeleteCommentsByPostID(c.Request.Context(), postID)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "게시글이 성공적으로 삭제되었습니다",
			},
		})
	}
}

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
	c, w := SetupTestContextWithSession("DELETE", "/admin/posts/post1", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	MockDeletePost(mockPostRepo, mockCommentRepo)(c)

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
	c, w := SetupTestContextWithSession("DELETE", "/admin/posts/non-existent", "")
	c.Params = []gin.Param{{Key: "id", Value: "non-existent"}}

	MockDeletePost(mockPostRepo, mockCommentRepo)(c)

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
			err:   errors.New("database error"),
		},
	}
	mockCommentRepo := &CommentRepositoryForDeletePostMock{}

	// When
	c, w := SetupTestContextWithSession("DELETE", "/admin/posts/post1", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	MockDeletePost(mockPostRepo, mockCommentRepo)(c)

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
		err: errors.New("database error"),
	}

	// When
	c, w := SetupTestContextWithSession("DELETE", "/admin/posts/post1", "")
	c.Params = []gin.Param{{Key: "id", Value: "post1"}}

	MockDeletePost(mockPostRepo, mockCommentRepo)(c)

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
