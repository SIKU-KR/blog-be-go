package main

import (
	"context"
	"encoding/gob"
	"log"
	"os"
	"time"

	_ "bumsiku/docs" // Swagger 문서 가져오기
	"bumsiku/internal/config"
	"bumsiku/internal/container"
	"bumsiku/internal/controller"
)

// @title           Bumsiku API
// @version         1.0
// @description     블로그 백엔드 API 서버
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey AdminAuth
// @in cookie
// @name loginSession
// @description 관리자 인증 세션 쿠키

func main() {
	if os.Getenv("GO_ENV") != "production" {
		config.LoadEnv()
	}

	gob.Register(time.Time{})

	ctx := context.Background()
	container, err := container.NewContainer(ctx)
	if err != nil {
		log.Fatalf("의존성 컨테이너 초기화 실패: %v", err)
	}

	r := controller.SetupRouter(container)

	if err := r.Run(); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
