package handler

import (
	"bumsiku/internal/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse는 모든 API 응답에 사용되는 공통 구조체입니다.
type APIResponse struct {
	Success bool        `json:"success" example:"true"` // 요청 성공 여부
	Data    interface{} `json:"data,omitempty"`         // 실제 데이터 (성공 시)
	Error   *APIError   `json:"error,omitempty"`        // 오류 정보 (실패 시)
}

// APIError는 API 오류 정보를 담는 구조체입니다.
type APIError struct {
	Code    string `json:"code" example:"BAD_REQUEST"`  // 오류 코드
	Message string `json:"message" example:"잘못된 요청입니다"` // 오류 메시지
}

// ErrorResponse Swagger 문서용 오류 응답 구조체
type ErrorResponse struct {
	Success bool     `json:"success" example:"false"` // 요청 성공 여부 (항상 false)
	Error   APIError `json:"error"`                   // 오류 정보
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

// SendErrorWithLogging은 오류를 로깅하고 응답을 반환하는 헬퍼 함수입니다.
func SendErrorWithLogging(c *gin.Context, logger *utils.Logger, statusCode int, errorCode string, message string, err error, contextInfo map[string]string) {
	// 요청 ID 가져오기
	requestID, exists := c.Get("requestID")
	requestIDStr := ""
	if exists {
		requestIDStr = fmt.Sprintf("%v", requestID)
	}

	// 기본 로그 필드 설정
	fields := map[string]string{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"statusCode": fmt.Sprintf("%d", statusCode),
		"errorCode":  errorCode,
	}

	// 컨텍스트 정보 추가
	for k, v := range contextInfo {
		fields[k] = v
	}

	// 요청 ID 추가
	if requestIDStr != "" {
		fields["requestID"] = requestIDStr
	}

	// 원본 오류가 있으면 추가
	errorDetail := message
	if err != nil {
		errorDetail = fmt.Sprintf("%s: %v", message, err)
		fields["errorDetail"] = err.Error()
	}

	// 오류 로깅
	logger.Error(context.Background(), errorDetail, fields)

	// 클라이언트에 응답
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

// SendBadRequestErrorWithLogging는 잘못된 요청 오류를 로깅하고 반환하는 헬퍼 함수입니다.
func SendBadRequestErrorWithLogging(c *gin.Context, logger *utils.Logger, message string, err error, contextInfo map[string]string) {
	SendErrorWithLogging(c, logger, http.StatusBadRequest, "BAD_REQUEST", message, err, contextInfo)
}

// SendNotFoundError는 리소스를 찾을 수 없음(404) 오류를 반환하는 헬퍼 함수입니다.
func SendNotFoundError(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// SendNotFoundErrorWithLogging는 리소스를 찾을 수 없음 오류를 로깅하고 반환하는 헬퍼 함수입니다.
func SendNotFoundErrorWithLogging(c *gin.Context, logger *utils.Logger, message string, err error, contextInfo map[string]string) {
	SendErrorWithLogging(c, logger, http.StatusNotFound, "NOT_FOUND", message, err, contextInfo)
}

// SendInternalServerError는 서버 내부 오류(500)를 반환하는 헬퍼 함수입니다.
func SendInternalServerError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// SendInternalServerErrorWithLogging는 서버 내부 오류를 로깅하고 반환하는 헬퍼 함수입니다.
func SendInternalServerErrorWithLogging(c *gin.Context, logger *utils.Logger, message string, err error, contextInfo map[string]string) {
	SendErrorWithLogging(c, logger, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, err, contextInfo)
}

// SendUnauthorizedError는 인증 실패(401) 오류를 반환하는 헬퍼 함수입니다.
func SendUnauthorizedError(c *gin.Context, message string) {
	SendError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// SendUnauthorizedErrorWithLogging는 인증 실패 오류를 로깅하고 반환하는 헬퍼 함수입니다.
func SendUnauthorizedErrorWithLogging(c *gin.Context, logger *utils.Logger, message string, err error, contextInfo map[string]string) {
	SendErrorWithLogging(c, logger, http.StatusUnauthorized, "UNAUTHORIZED", message, err, contextInfo)
}

// SendForbiddenError는 권한 없음(403) 오류를 반환하는 헬퍼 함수입니다.
func SendForbiddenError(c *gin.Context, message string) {
	SendError(c, http.StatusForbidden, "FORBIDDEN", message)
}

// SendForbiddenErrorWithLogging는 권한 없음 오류를 로깅하고 반환하는 헬퍼 함수입니다.
func SendForbiddenErrorWithLogging(c *gin.Context, logger *utils.Logger, message string, err error, contextInfo map[string]string) {
	SendErrorWithLogging(c, logger, http.StatusForbidden, "FORBIDDEN", message, err, contextInfo)
}
