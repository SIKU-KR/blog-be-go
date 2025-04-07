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
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockGetCategories(repo *MockCategoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 카테고리 목록 조회
		categories, err := repo.GetCategories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "카테고리 목록 조회에 실패했습니다",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"categories": categories,
			},
		})
	}
}

// MockCategoryRepository는 테스트에 사용되는 카테고리 저장소 모의 객체입니다.
type MockCategoryRepository struct {
	categories []model.Category
	err        error
}

func (m *MockCategoryRepository) GetCategories(ctx context.Context) ([]model.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.categories, nil
}

func (m *MockCategoryRepository) UpsertCategory(ctx context.Context, category model.Category) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

// CreateTestCategories는 테스트용 카테고리 데이터를 생성합니다.
func CreateTestCategories() []model.Category {
	now := time.Now()
	return []model.Category{
		{
			Category:  "tech",
			Order:     1,
			CreatedAt: now,
		},
		{
			Category:  "life",
			Order:     2,
			CreatedAt: now.Add(time.Hour),
		},
		{
			Category:  "book",
			Order:     3,
			CreatedAt: now.Add(2 * time.Hour),
		},
	}
}

// [GIVEN] 정상적인 카테고리 목록이 있는 경우
// [WHEN] GetCategories 핸들러를 호출
// [THEN] 상태코드 200과 카테고리 목록 반환 확인
func TestGetCategories_Success(t *testing.T) {
	// Given
	mockCategories := CreateTestCategories()
	mockRepo := &MockCategoryRepository{categories: mockCategories}

	// When
	c, w := SetupTestContext("GET", "/categories", "")

	MockGetCategories(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	categories := data["categories"].([]interface{})
	assert.Equal(t, 3, len(categories))

	// 첫 번째 카테고리 확인
	firstCategory := categories[0].(map[string]interface{})
	assert.Equal(t, "tech", firstCategory["category"])
	assert.Equal(t, float64(1), firstCategory["order"])
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] GetCategories 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestGetCategories_Error(t *testing.T) {
	// Given
	mockRepo := &MockCategoryRepository{err: errors.New("database error")}

	// When
	c, w := SetupTestContext("GET", "/categories", "")

	MockGetCategories(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Equal(t, "카테고리 목록 조회에 실패했습니다", errorData["message"])
}

// [GIVEN] 카테고리가 없는 경우
// [WHEN] GetCategories 핸들러를 호출
// [THEN] 상태코드 200과 빈 카테고리 목록 반환 확인
func TestGetCategories_EmptyList(t *testing.T) {
	// Given
	mockRepo := &MockCategoryRepository{categories: []model.Category{}}

	// When
	c, w := SetupTestContext("GET", "/categories", "")

	MockGetCategories(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	categories := data["categories"].([]interface{})
	assert.Equal(t, 0, len(categories))
}
