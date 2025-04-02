package container

import (
	"bumsiku/internal/repository"
	"bumsiku/pkg/client"
	"context"
)

type Container struct {
	PostRepository    *repository.PostRepository
	CommentRepository *repository.CommentRepository
}

func NewContainer(ctx context.Context) (*Container, error) {
	ddbClient, err := client.NewDdbClient(ctx)
	if err != nil {
		return nil, err
	}

	postRepo := repository.NewPostRepository(ddbClient)
	commentRepo := repository.NewCommentRepository(ddbClient)

	return &Container{
		PostRepository:    postRepo,
		CommentRepository: commentRepo,
	}, nil
}
