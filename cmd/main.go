package main

import (
	"context"
	"encoding/gob"
	"log"
	"os"
	"time"

	"bumsiku/internal/config"
	"bumsiku/internal/container"
	"bumsiku/internal/controller"
)

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
