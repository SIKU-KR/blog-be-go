package middleware

import (
	"bumsiku/internal/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// bodyLogWriter 응답 바디를 저장하는 커스텀 gin.ResponseWriter
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 원본 ResponseWriter의 Write를 오버라이드하여 응답 바디를 캡처
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware 모든 요청과 응답을 로깅하는 미들웨어
func LoggingMiddleware(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 요청 시작 시간
		startTime := time.Now()

		// 요청 ID 생성
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		// 요청 바디 읽기
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Body를 다시 설정해야 함 (한 번 읽으면 내용이 소진됨)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 응답 바디를 캡처하기 위한 커스텀 writer
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 요청 로깅
		fields := map[string]string{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"ip":        c.ClientIP(),
			"userAgent": c.Request.UserAgent(),
			"requestID": requestID,
		}

		// 요청 바디 로깅 (선택적으로 사용 가능)
		if len(requestBody) > 0 && shouldLogBody(c.Request.URL.Path) {
			// 보안 상 로깅하면 안 되는 필드 필터링 (비밀번호 등)
			fields["requestBody"] = filterSensitiveData(string(requestBody))
		}

		logger.Info(context.Background(), fmt.Sprintf("요청 시작: %s %s", c.Request.Method, c.Request.URL.Path), fields)

		// 다음 핸들러 호출
		c.Next()

		// 응답 시간 계산
		duration := time.Since(startTime)

		// 응답 로깅
		responseFields := map[string]string{
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"requestID":     requestID,
			"statusCode":    fmt.Sprintf("%d", c.Writer.Status()),
			"duration":      duration.String(),
			"contentType":   c.Writer.Header().Get("Content-Type"),
			"contentLength": fmt.Sprintf("%d", c.Writer.Size()),
		}

		// 응답 바디 (선택적으로 사용 가능)
		if shouldLogBody(c.Request.URL.Path) && blw.body.Len() > 0 {
			responseBody := blw.body.String()
			// 보안 상 로깅하면 안 되는 필드 필터링
			responseFields["responseBody"] = filterSensitiveData(responseBody)
		}

		// 에러가 있었을 경우 로깅 레벨 변경
		logLevel := utils.LogLevelInfo
		if c.Writer.Status() >= 400 {
			logLevel = utils.LogLevelError
			// 에러 내용 추가
			if len(c.Errors) > 0 {
				responseFields["errors"] = c.Errors.String()
			}
		}

		// 응답 로깅
		logger.Log(
			context.Background(),
			logLevel,
			fmt.Sprintf("응답 완료: %s %s %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status()),
			responseFields,
		)
	}
}

// shouldLogBody 해당 경로의 요청/응답 바디를 로깅해야 하는지 결정
func shouldLogBody(path string) bool {
	// 필요에 따라 특정 경로의 바디를 로깅하지 않을 수 있음
	// 예: 파일 업로드 같은 대용량 데이터
	if path == "/admin/images" || path == "/swagger/*any" {
		return false
	}
	return true
}

// filterSensitiveData 민감한 데이터를 필터링 (비밀번호 등)
func filterSensitiveData(data string) string {
	// 여기서 정규식 등을 사용하여 비밀번호, 토큰 등 민감 데이터 필터링 가능
	// 간단한 예제: 실제 구현시 정규식으로 더 정확하게 처리해야 함
	return data
}
