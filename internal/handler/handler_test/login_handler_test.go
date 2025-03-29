package handler_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"bumsiku/internal/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

// [GIVEN] 올바른 자격증명을 포함한 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 200과 "Login Successful" 메시지 반환 확인
func TestPostLogin_Success(t *testing.T) {
	givenTestEnv()
	body := `{"username": "admin", "password": "password"}`
	w := whenPostLogin(body)
	thenAssertResponseJSON(t, w, http.StatusOK, "message", "Login Successful")
}

// [GIVEN] 잘못된 자격증명을 포함한 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 401과 "Failed to login" 에러 메시지 반환 확인
func TestPostLogin_InvalidCredentials(t *testing.T) {
	givenTestEnv()
	body := `{"username": "wrong", "password": "creds"}`
	w := whenPostLogin(body)
	thenAssertResponseJSON(t, w, http.StatusUnauthorized, "error", "Failed to login")
}

// [GIVEN] 필수 필드가 누락된 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 400과 "Bad Request" 에러 메시지 반환 확인
func TestPostLogin_BadRequest(t *testing.T) {
	givenTestEnv()
	body := `{"username": "admin"}`
	w := whenPostLogin(body)
	thenAssertResponseJSON(t, w, http.StatusBadRequest, "error", "Bad Request")
}

func init() {
	gob.Register(time.Time{})
}

// 테스트용 gin.Context와 ResponseRecorder를 생성합니다.
func givenTestContext(method, url, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	// 세션 미들웨어를 통해 세션 객체를 컨텍스트에 등록합니다.
	store := memstore.NewStore([]byte("secret"))
	sessionsMiddleware := sessions.Sessions("mysession", store)
	sessionsMiddleware(c)

	return c, w
}

// whenPostLogin은 주어진 요청 본문(body)으로 PostLogin 핸들러를 호출합니다.
func whenPostLogin(body string) *httptest.ResponseRecorder {
	c, w := givenTestContext("POST", "/login", body)
	handler.PostLogin(c)
	return w
}

// 테스트에 필요한 환경변수를 설정
func givenTestEnv() {
	os.Setenv("ADMIN_ID", "admin")
	os.Setenv("ADMIN_PW", "password")
	gin.SetMode(gin.TestMode)
}

// 예상 상태코드와 응답 메시지(또는 에러)를 검증
func thenAssertResponseJSON(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedKey, expectedValue string) {
	if w.Code != expectedStatus {
		t.Fatalf("상태 코드 %d 기대했으나 %d 반환됨", expectedStatus, w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("응답 파싱 실패: %v", err)
	}
	if resp[expectedKey] != expectedValue {
		t.Errorf("'%s'의 값이 '%s' 이어야 하는데, 반환: %v", expectedKey, expectedValue, resp[expectedKey])
	}
}
