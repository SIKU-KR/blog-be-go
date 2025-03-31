package container

import (
	"bumsiku/internal/repository"
	"bumsiku/pkg/client"
	"context"
)

type Container struct {
	PostRepository *repository.PostRepository
}

func NewContainer(ctx context.Context) (*Container, error) {
	ddbClient, err := client.NewDdbClient(ctx)
	if err != nil {
		return nil, err
	}

	postRepo := repository.NewPostRepository(ddbClient)

	return &Container{
		PostRepository: postRepo,
	}, nil
}
