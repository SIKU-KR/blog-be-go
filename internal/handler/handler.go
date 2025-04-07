package handler

import (
	"bumsiku/internal/utils"
)

// Handler는 모든 핸들러에서 공통으로 사용하는 의존성을 관리하는 구조체입니다.
type Handler struct {
	Logger *utils.Logger
}

// NewHandler는 새로운 핸들러 인스턴스를 생성합니다.
func NewHandler(logger *utils.Logger) *Handler {
	return &Handler{
		Logger: logger,
	}
}
