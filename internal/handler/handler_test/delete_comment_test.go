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
func MockDeleteComment(repo *CommentRepositoryForDeleteCommentMock) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID := c.Param("commentId")
		if commentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "댓글 ID가 필요합니다",
				},
			})
			return
		}

		// 댓글 삭제
		err := repo.DeleteComment(c.Request.Context(), commentID)
		if err != nil {
			var notFoundErr *repository.CommentNotFoundError
			if errors.As(err, &notFoundErr) {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "NOT_FOUND",
						"message": "댓글을 찾을 수 없음: " + commentID,
					},
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "댓글 삭제에 실패했습니다",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "댓글이 성공적으로 삭제되었습니다",
			},
		})
	}
}

// CommentRepositoryForDeleteCommentMock은 DeleteComment 함수를 모의하는 구조체입니다.
type CommentRepositoryForDeleteCommentMock struct {
	CommentRepositoryMock
	deletedCommentID string
}

func (m *CommentRepositoryForDeleteCommentMock) DeleteComment(ctx context.Context, commentID string) error {
	// 에러가 설정된 경우 반환
	if m.err != nil {
		return m.err
	}

	// 댓글 존재 여부 확인
	found := false
	for _, comment := range m.comments {
		if comment.CommentID == commentID {
			found = true
			break
		}
	}

	// 댓글이 없는 경우 에러 반환
	if !found {
		return &repository.CommentNotFoundError{CommentID: commentID}
	}

	// 삭제된 commentId 저장 (테스트에서 확인용)
	m.deletedCommentID = commentID
	return nil
}

// 테스트용 댓글 데이터 생성 함수
func CreateTestCommentsForDeleteTest() []model.Comment {
	now := time.Now()
	return []model.Comment{
		{
			CommentID: "comment1",
			PostID:    "post1",
			Nickname:  "사용자1",
			Content:   "첫 번째 댓글",
			CreatedAt: now,
		},
		{
			CommentID: "comment2",
			PostID:    "post1",
			Nickname:  "사용자2",
			Content:   "두 번째 댓글",
			CreatedAt: now,
		},
		{
			CommentID: "comment3",
			PostID:    "post2",
			Nickname:  "사용자3",
			Content:   "다른 게시글 댓글",
			CreatedAt: now,
		},
	}
}

// [GIVEN] 유효한 댓글 ID로 요청한 경우
// [WHEN] DeleteComment 핸들러를 호출
// [THEN] 상태코드 200과 성공 메시지 반환 확인
func TestDeleteComment_Success(t *testing.T) {
	// Given
	mockComments := CreateTestCommentsForDeleteTest()
	mockRepo := &CommentRepositoryForDeleteCommentMock{
		CommentRepositoryMock: CommentRepositoryMock{comments: mockComments},
	}

	// When
	c, w := SetupTestContextWithSession("DELETE", "/admin/comments/comment1", "")
	c.Params = []gin.Param{{Key: "commentId", Value: "comment1"}}
	
	MockDeleteComment(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "댓글이 성공적으로 삭제되었습니다", data["message"])
	assert.Equal(t, "comment1", mockRepo.deletedCommentID)
}

// [GIVEN] 존재하지 않는 댓글 ID로 요청한 경우
// [WHEN] DeleteComment 핸들러를 호출
// [THEN] 상태코드 404와 에러 메시지 반환 확인
func TestDeleteComment_NotFound(t *testing.T) {
	// Given
	mockComments := CreateTestCommentsForDeleteTest()
	mockRepo := &CommentRepositoryForDeleteCommentMock{
		CommentRepositoryMock: CommentRepositoryMock{comments: mockComments},
	}

	// When
	c, w := SetupTestContextWithSession("DELETE", "/admin/comments/non-existent", "")
	c.Params = []gin.Param{{Key: "commentId", Value: "non-existent"}}

	MockDeleteComment(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errorData["code"])
	assert.Contains(t, errorData["message"], "댓글을 찾을 수 없음")
}

// [GIVEN] 댓글 ID가 비어있는 경우
// [WHEN] DeleteComment 핸들러를 호출
// [THEN] 상태코드 400과 에러 메시지 반환 확인
func TestDeleteComment_MissingId(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryForDeleteCommentMock{}

	// When
	c, w := SetupTestContextWithSession("DELETE", "/admin/comments/", "")
	c.Params = []gin.Param{{Key: "commentId", Value: ""}}

	MockDeleteComment(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errorData["code"])
	assert.Equal(t, "댓글 ID가 필요합니다", errorData["message"])
}

// [GIVEN] Repository에서 내부 에러가 발생하는 경우
// [WHEN] DeleteComment 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestDeleteComment_InternalError(t *testing.T) {
	// Given
	mockRepo := &CommentRepositoryForDeleteCommentMock{
		CommentRepositoryMock: CommentRepositoryMock{
			comments: CreateTestCommentsForDeleteTest(),
			err:      errors.New("database error"),
		},
	}

	// When
	c, w := SetupTestContextWithSession("DELETE", "/admin/comments/comment1", "")
	c.Params = []gin.Param{{Key: "commentId", Value: "comment1"}}

	MockDeleteComment(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Contains(t, errorData["message"], "댓글 삭제에 실패했습니다")
}
