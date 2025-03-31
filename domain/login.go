package domain

// LoginRequest는 사용자 로그인 요청 데이터를 담는 구조체입니다.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
