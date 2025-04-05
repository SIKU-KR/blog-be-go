package container

import (
	"bumsiku/internal/repository"
	"bumsiku/pkg/client"
	"context"
)

type Container struct {
	PostRepository     *repository.PostRepository
	CommentRepository  *repository.CommentRepository
	CategoryRepository *repository.CategoryRepository
}

func NewContainer(ctx context.Context) (*Container, error) {
	ddbClient, err := client.NewDdbClient(ctx)
	if err != nil {
		return nil, err
	}

	postRepo := repository.NewPostRepository(ddbClient)
	commentRepo := repository.NewCommentRepository(ddbClient)
	categoryRepo := repository.NewCategoryRepository(ddbClient)

	return &Container{
		PostRepository:     postRepo,
		CommentRepository:  commentRepo,
		CategoryRepository: categoryRepo,
	}, nil
}
