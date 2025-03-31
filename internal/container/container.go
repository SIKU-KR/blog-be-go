package container

import (
	"bumsiku/internal/client"
	"bumsiku/internal/repository"
	"context"
)

type Container struct {
	PostRepository *repository.PostRepository
}

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