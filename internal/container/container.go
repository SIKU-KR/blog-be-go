package container

import (
	"bumsiku/internal/repository"
	"bumsiku/pkg/client"
	"context"
)

// Container는 애플리케이션의 모든 의존성을 관리하는 구조체입니다.
type Container struct {
	PostRepository *repository.PostRepository
}

// NewContainer는 새로운 Container 인스턴스를 생성하고 초기화합니다.
func NewContainer(ctx context.Context) (*Container, error) {
	// DynamoDB 클라이언트 초기화
	ddbClient, err := client.NewDdbClient(ctx)
	if err != nil {
		return nil, err
	}

	// Repository 초기화
	postRepo := repository.NewPostRepository(ddbClient)

	return &Container{
		PostRepository: postRepo,
	}, nil
}
