package handler

import (
	"bumsiku/internal/handler"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// [GIVEN] 유효한 카테고리 업데이트 요청이 있는 경우
// [WHEN] UpdateCategory 핸들러를 호출
// [THEN] 상태코드 200과 업데이트된 카테고리 정보 반환 확인
func TestUpdateCategory_Success(t *testing.T) {
	// Given
	mockRepo := &MockCategoryRepository{}
	validRequest := `{
		"category": "tech",
		"order": 1
	}`

	// When
	c, w := SetupTestContext("PUT", "/admin/categories", validRequest)
	handler.UpdateCategory(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "tech", data["category"])
	assert.Equal(t, float64(1), data["order"])
}

// [GIVEN] 유효하지 않은 요청 형식(필수 필드 누락)이 있는 경우
// [WHEN] UpdateCategory 핸들러를 호출
// [THEN] 상태코드 400과 에러 메시지 반환 확인
func TestUpdateCategory_InvalidRequest_MissingField(t *testing.T) {
	// Given
	mockRepo := &MockCategoryRepository{}
	invalidRequest := `{
		"category": "tech"
	}`

	// When
	c, w := SetupTestContext("PUT", "/admin/categories", invalidRequest)
	handler.UpdateCategory(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errorData["code"])
	assert.Contains(t, errorData["message"].(string), "요청 형식이 올바르지 않습니다")
}

// [GIVEN] 유효하지 않은 JSON 형식의 요청이 있는 경우
// [WHEN] UpdateCategory 핸들러를 호출
// [THEN] 상태코드 400과 에러 메시지 반환 확인
func TestUpdateCategory_InvalidRequest_MalformedJSON(t *testing.T) {
	// Given
	mockRepo := &MockCategoryRepository{}
	invalidJSON := `{
		"category": "tech",
		"order": 1,
	}` // 잘못된 JSON 형식 (trailing comma)

	// When
	c, w := SetupTestContext("PUT", "/admin/categories", invalidJSON)
	handler.UpdateCategory(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errorData["code"])
	assert.Contains(t, errorData["message"].(string), "요청 형식이 올바르지 않습니다")
}

// [GIVEN] Repository에서 에러가 발생하는 경우
// [WHEN] UpdateCategory 핸들러를 호출
// [THEN] 상태코드 500과 에러 메시지 반환 확인
func TestUpdateCategory_RepositoryError(t *testing.T) {
	// Given
	mockRepo := &MockCategoryRepository{err: assert.AnError}
	validRequest := `{
		"category": "tech",
		"order": 1
	}`

	// When
	c, w := SetupTestContext("PUT", "/admin/categories", validRequest)
	handler.UpdateCategory(mockRepo)(c)

	// Then
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorData["code"])
	assert.Contains(t, errorData["message"].(string), "카테고리 업데이트에 실패했습니다")
}
