package repository

import (
	"context"

	"bumsiku/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const CommentTableName = "blog_comments"

type CommentRepositoryInterface interface {
	GetComments(ctx context.Context, input *GetCommentsInput) ([]model.Comment, error)
}

type CommentRepository struct {
	client *dynamodb.Client
}

func NewCommentRepository(client *dynamodb.Client) *CommentRepository {
	return &CommentRepository{client: client}
}

type GetCommentsInput struct {
	PostID *string
}

func (r *CommentRepository) GetComments(ctx context.Context, input *GetCommentsInput) ([]model.Comment, error) {
	var comments []model.Comment
	var err error

	if input.PostID != nil && *input.PostID != "" {
		// 특정 게시글의 댓글만 조회
		comments, err = r.getCommentsByPostID(ctx, *input.PostID)
	} else {
		// 모든 댓글 조회
		comments, err = r.getAllComments(ctx)
	}

	if err != nil {
		return nil, err
	}

	return comments, nil
}

// 특정 게시글의 댓글 조회
func (r *CommentRepository) getCommentsByPostID(ctx context.Context, postID string) ([]model.Comment, error) {
	keyCondition := expression.Key("postId").Equal(expression.Value(postID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, err
	}

	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(CommentTableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(true), // 등록순 정렬
	})

	if err != nil {
		return nil, err
	}

	// 결과 변환
	comments := make([]model.Comment, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// 모든 댓글 조회
func (r *CommentRepository) getAllComments(ctx context.Context) ([]model.Comment, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(CommentTableName),
	})

	if err != nil {
		return nil, err
	}

	// 결과 변환
	comments := make([]model.Comment, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}
