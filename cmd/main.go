package main

import (
	"context"
	"encoding/gob"
	"log"
	"os"
	"time"

	"bumsiku/controller"
	"bumsiku/internal/config"
	"bumsiku/internal/container"
)

func main() {
	if os.Getenv("GO_ENV") != "production" {
		config.LoadEnv()
	}

	gob.Register(time.Time{})

	// 의존성 컨테이너 초기화
	ctx := context.Background()
	container, err := container.NewContainer(ctx)
	if err != nil {
		log.Fatalf("의존성 컨테이너 초기화 실패: %v", err)
	}

	// 라우터 설정
	r := controller.SetupRouter(container)

	// 서버 시작
	if err := r.Run(); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
