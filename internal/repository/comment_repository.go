package repository

import (
	"context"
	"time"

	"bumsiku/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const CommentTableName = "blog_comments"

type CommentRepositoryInterface interface {
	GetComments(ctx context.Context, input *GetCommentsInput) ([]model.Comment, error)
	CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	DeleteCommentsByPostID(ctx context.Context, postID string) error
	DeleteComment(ctx context.Context, commentID string) error
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

// CreateComment는 댓글을 생성합니다
func (r *CommentRepository) CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	comment.CommentID = uuid.New().String()
	comment.CreatedAt = time.Now()

	// DynamoDB 아이템으로 변환
	item, err := attributevalue.MarshalMap(comment)
	if err != nil {
		return nil, err
	}

	// DynamoDB에 삽입
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(CommentTableName),
		Item:      item,
	})

	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *CommentRepository) DeleteCommentsByPostID(ctx context.Context, postID string) error {
	// 먼저 해당 게시글의 모든 댓글을 조회합니다
	comments, err := r.getCommentsByPostID(ctx, postID)
	if err != nil {
		return err
	}

	// 댓글이 없으면 바로 성공 반환
	if len(comments) == 0 {
		return nil
	}

	// 배치 쓰기 요청을 준비합니다
	var writeRequests []types.WriteRequest
	for _, comment := range comments {
		writeRequests = append(writeRequests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"postId":    &types.AttributeValueMemberS{Value: comment.PostID},
					"commentId": &types.AttributeValueMemberS{Value: comment.CommentID},
				},
			},
		})
	}

	// DynamoDB의 BatchWriteItem은 한 번에 최대 25개 항목만 처리할 수 있으므로
	// 25개씩 나누어 처리합니다
	for i := 0; i < len(writeRequests); i += 25 {
		end := i + 25
		if end > len(writeRequests) {
			end = len(writeRequests)
		}

		batch := writeRequests[i:end]
		_, err := r.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				CommentTableName: batch,
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// CommentNotFoundError는 댓글을 찾을 수 없을 때 발생하는 오류입니다.
type CommentNotFoundError struct {
	CommentID string
}

func (e *CommentNotFoundError) Error() string {
	return "댓글을 찾을 수 없음: " + e.CommentID
}

// DeleteComment는 특정 댓글을 삭제합니다.
func (r *CommentRepository) DeleteComment(ctx context.Context, commentID string) error {
	// 먼저 댓글이 존재하는지 확인
	// 모든 댓글을 조회하여 찾아야 함 (DynamoDB에서 commentId로 직접 찾을 수 없기 때문)
	allComments, err := r.getAllComments(ctx)
	if err != nil {
		return err
	}

	// 댓글 검색
	var targetComment *model.Comment
	for _, comment := range allComments {
		if comment.CommentID == commentID {
			targetComment = &comment
			break
		}
	}

	// 댓글이 존재하지 않는 경우
	if targetComment == nil {
		return &CommentNotFoundError{CommentID: commentID}
	}

	// 댓글 삭제
	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(CommentTableName),
		Key: map[string]types.AttributeValue{
			"postId":    &types.AttributeValueMemberS{Value: targetComment.PostID},
			"commentId": &types.AttributeValueMemberS{Value: targetComment.CommentID},
		},
	})

	return err
}
