package handler

import (
	"bumsiku/domain"
	"bumsiku/internal/handler"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// [GIVEN] 정상적인 게시글 목록이 있는 경우
// [WHEN] GetPosts 핸들러를 호출
// [THEN] 상태코드 200과 게시글 목록 반환 확인
func TestGetPosts_Success(t *testing.T) {
	// Given
	mockPosts := CreateTestPosts()
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts", "")
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
	mockPosts := []domain.Post{CreateTestPosts()[0]} // tech 카테고리만
	mockRepo := &mockPostRepository{posts: mockPosts}

	// When
	c, w := SetupTestContext("GET", "/posts?category=tech", "")
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
	nextToken := "next_page_token"
	mockRepo := &mockPostRepository{
		posts:     CreateTestPosts()[:1],
		nextToken: &nextToken,
	}

	// When
	c, w := SetupTestContext("GET", "/posts?pageSize=1", "")
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
	mockRepo := &mockPostRepository{err: assert.AnError}

	// When
	c, w := SetupTestContext("GET", "/posts", "")
	handler.GetPosts(mockRepo)(c)

	// Then
	AssertResponseJSON(t, w, http.StatusInternalServerError, "error", "게시글 목록 조회에 실패했습니다")
}
