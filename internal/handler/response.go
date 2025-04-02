package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse는 모든 API 응답에 사용되는 공통 구조체입니다.
type APIResponse struct {
	Success bool        `json:"success"`      // 요청 성공 여부
	Data    interface{} `json:"data,omitempty"` // 실제 데이터 (성공 시)
	Error   *APIError   `json:"error,omitempty"` // 오류 정보 (실패 시)
}

// APIError는 API 오류 정보를 담는 구조체입니다.
type APIError struct {
	Code    string `json:"code"`              // 오류 코드
	Message string `json:"message"`           // 오류 메시지
}

// SendSuccess는 성공 응답을 반환하는 헬퍼 함수입니다.
func SendSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// SendError는 오류 응답을 반환하는 헬퍼 함수입니다.
func SendError(c *gin.Context, statusCode int, errorCode string, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    errorCode,
			Message: message,
		},
	})
}

// SendBadRequestError는 잘못된 요청(400) 오류를 반환하는 헬퍼 함수입니다.
func SendBadRequestError(c *gin.Context, message string) {
	SendError(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// SendNotFoundError는 리소스를 찾을 수 없음(404) 오류를 반환하는 헬퍼 함수입니다.
func SendNotFoundError(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// SendInternalServerError는 서버 내부 오류(500)를 반환하는 헬퍼 함수입니다.
func SendInternalServerError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// SendUnauthorizedError는 인증 실패(401) 오류를 반환하는 헬퍼 함수입니다.
func SendUnauthorizedError(c *gin.Context, message string) {
	SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// SendForbiddenError는 권한 없음(403) 오류를 반환하는 헬퍼 함수입니다.
func SendForbiddenError(c *gin.Context, message string) {
	SendError(c, http.StatusForbidden, "FORBIDDEN", message)
} 