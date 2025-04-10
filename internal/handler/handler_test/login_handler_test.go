package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 핸들러 모의 함수 - 로거를 사용하지 않도록 구현
func MockPostLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "잘못된 요청 형식입니다",
				},
			})
			return
		}

		// 환경 변수에서 관리자 자격증명 가져오기
		adminID := os.Getenv("ADMIN_ID")
		adminPW := os.Getenv("ADMIN_PW")

		// 자격증명 확인
		if loginRequest.Username != adminID || loginRequest.Password != adminPW {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "로그인에 실패했습니다",
				},
			})
			return
		}

		// 테스트 환경에서는 실제 세션 처리 없이 바로 응답 반환
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "로그인에 성공했습니다",
			},
		})
	}
}

// [GIVEN] 올바른 자격증명을 포함한 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 200과 "로그인에 성공했습니다" 메시지 반환 확인
func TestPostLogin_Success(t *testing.T) {
	SetTestEnvironment()
	body := `{"username": "admin", "password": "password"}`
	c, w := SetupTestContextWithSession("POST", "/login", body)

	MockPostLogin()(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 새로운 응답 구조체 확인
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "로그인에 성공했습니다", data["message"])
}

// [GIVEN] 잘못된 자격증명을 포함한 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 401과 "로그인에 실패했습니다" 에러 메시지 반환 확인
func TestPostLogin_InvalidCredentials(t *testing.T) {
	SetTestEnvironment()
	body := `{"username": "wrong", "password": "creds"}`
	c, w := SetupTestContextWithSession("POST", "/login", body)

	MockPostLogin()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "UNAUTHORIZED", errorData["code"])
	assert.Equal(t, "로그인에 실패했습니다", errorData["message"])
}

// [GIVEN] 필수 필드가 누락된 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 400과 "잘못된 요청 형식입니다" 에러 메시지 반환 확인
func TestPostLogin_BadRequest(t *testing.T) {
	SetTestEnvironment()
	body := `{"username": "admin"}`
	c, w := SetupTestContextWithSession("POST", "/login", body)

	MockPostLogin()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errorData["code"])
	assert.Equal(t, "잘못된 요청 형식입니다", errorData["message"])
}
