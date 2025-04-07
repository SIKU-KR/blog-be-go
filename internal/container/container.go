package container

import (
	"bumsiku/internal/repository"
	"bumsiku/pkg/client"
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Container struct {
	PostRepository     *repository.PostRepository
	CommentRepository  *repository.CommentRepository
	CategoryRepository *repository.CategoryRepository
	S3Client           *s3.Client
	CloudWatchClient   *cloudwatchlogs.Client
}

func NewContainer(ctx context.Context) (*Container, error) {
	ddbClient, err := client.NewDdbClient(ctx)
	if err != nil {
		return nil, err
	}

	s3Client, err := client.NewS3Client(ctx)
	if err != nil {
		return nil, err
	}

	cwClient, err := client.NewCloudWatchLogsClient(ctx)
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
		S3Client:           s3Client,
		CloudWatchClient:   cwClient,
	}, nil
}
