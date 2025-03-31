package handler

import (
	"bumsiku/internal/handler"
	"net/http"
	"testing"
)

// [GIVEN] 올바른 자격증명을 포함한 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 200과 "Login Successful" 메시지 반환 확인
func TestPostLogin_Success(t *testing.T) {
	SetTestEnvironment()
	body := `{"username": "admin", "password": "password"}`
	c, w := SetupTestContextWithSession("POST", "/login", body)
	handler.PostLogin(c)
	AssertResponseJSON(t, w, http.StatusOK, "message", "Login Successful")
}

// [GIVEN] 잘못된 자격증명을 포함한 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 401과 "Failed to login" 에러 메시지 반환 확인
func TestPostLogin_InvalidCredentials(t *testing.T) {
	SetTestEnvironment()
	body := `{"username": "wrong", "password": "creds"}`
	c, w := SetupTestContextWithSession("POST", "/login", body)
	handler.PostLogin(c)
	AssertResponseJSON(t, w, http.StatusUnauthorized, "error", "Failed to login")
}

// [GIVEN] 필수 필드가 누락된 JSON 페이로드를 준비
// [WHEN] PostLogin 핸들러를 호출
// [THEN] 상태코드 400과 "Bad Request" 에러 메시지 반환 확인
func TestPostLogin_BadRequest(t *testing.T) {
	SetTestEnvironment()
	body := `{"username": "admin"}`
	c, w := SetupTestContextWithSession("POST", "/login", body)
	handler.PostLogin(c)
	AssertResponseJSON(t, w, http.StatusBadRequest, "error", "Bad Request")
}
