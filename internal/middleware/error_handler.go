package middleware

import (
	"bumsiku/internal/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 에러 응답 형식
type ErrorResponse struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
}

// ErrorHandlingMiddleware 에러 핸들링 미들웨어
func ErrorHandlingMiddleware(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 다음 핸들러 호출
		c.Next()

		// 에러가 있는지 확인
		if len(c.Errors) > 0 {
			// 첫 번째 에러 가져오기
			err := c.Errors.Last()
			requestID, _ := c.Get("requestID")

			// 상태 코드 설정
			status := http.StatusInternalServerError

			// gin.Error 타입에 따라 상태 코드 추정
			if err.Type == gin.ErrorTypePublic {
				status = http.StatusBadRequest
			} else if err.Type == gin.ErrorTypeBind {
				status = http.StatusBadRequest
			}

			// 이미 상태 코드가 설정되어 있다면 사용
			if c.Writer.Status() != http.StatusOK {
				status = c.Writer.Status()
			}

			// 에러 로깅
			fields := map[string]string{
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"ip":        c.ClientIP(),
				"errorType": fmt.Sprintf("%d", err.Type),
			}

			if requestID != nil {
				fields["requestID"] = requestID.(string)
			}

			// 에러 로깅
			logger.Error(context.Background(), fmt.Sprintf("에러 발생: %v", err.Error()), fields)

			// 에러 응답 보내기
			c.JSON(status, ErrorResponse{
				Status:    status,
				Message:   err.Error(),
				RequestID: fmt.Sprintf("%v", requestID),
			})
			c.Abort()
		}
	}
}

// RecoveryWithLogger panic을 복구하고 로깅하는 미들웨어
func RecoveryWithLogger(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get("requestID")

				// panic 로깅
				fields := map[string]string{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"ip":     c.ClientIP(),
				}

				if requestID != nil {
					fields["requestID"] = requestID.(string)
				}

				logger.Error(context.Background(), fmt.Sprintf("Panic 복구: %v", r), fields)

				// 클라이언트에게 500 에러 반환
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Status:    http.StatusInternalServerError,
					Message:   "서버 내부 오류가 발생했습니다",
					RequestID: fmt.Sprintf("%v", requestID),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
