package repository

import (
	"bumsiku/domain"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const POST_TABLE_NAME = "posts"
const PAGE_SIZE = 10

type PostRepositoryInterface interface {
	GetPosts(ctx context.Context, input *GetPostsInput) (*GetPostsOutput, error)
}

type PostRepository struct {
	client *dynamodb.Client
}

func NewPostRepository(client *dynamodb.Client) *PostRepository {
	return &PostRepository{client: client}
}

type GetPostsInput struct {
	Category    *string
	NextToken   *string
	PageSize    *int32
}

type GetPostsOutput struct {
	Posts     []domain.Post
	NextToken *string
}

func (r *PostRepository) GetPosts(ctx context.Context, input *GetPostsInput) (*GetPostsOutput, error) {
	// 기본 쿼리 설정
	keyEx := expression.Key("postId").BeginsWith("")
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, err
	}

	// 페이지 크기 설정
	limit := PAGE_SIZE
	if input.PageSize != nil {
		limit = int(*input.PageSize)
	}

	// 쿼리 입력 구성
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(POST_TABLE_NAME),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                    aws.Int32(int32(limit)),
		ScanIndexForward:         aws.Bool(false), // 최신순 정렬
	}

	// 카테고리 필터링이 있는 경우
	if input.Category != nil && *input.Category != "" {
		queryInput.IndexName = aws.String("category-index")
		keyEx = expression.Key("category").Equal(expression.Value(*input.Category))
		expr, err = expression.NewBuilder().WithKeyCondition(keyEx).Build()
		if err != nil {
			return nil, err
		}
		queryInput.KeyConditionExpression = expr.KeyCondition()
		queryInput.ExpressionAttributeNames = expr.Names()
		queryInput.ExpressionAttributeValues = expr.Values()
	}

	// 페이지네이션 토큰이 있는 경우
	if input.NextToken != nil {
		queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
			"postId": &types.AttributeValueMemberS{Value: *input.NextToken},
		}
	}

	// 쿼리 실행
	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	// 결과 변환
	posts := make([]domain.Post, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &posts)
	if err != nil {
		return nil, err
	}

	// 다음 페이지 토큰 설정
	var nextToken *string
	if result.LastEvaluatedKey != nil {
		if postId, ok := result.LastEvaluatedKey["postId"].(*types.AttributeValueMemberS); ok {
			nextToken = aws.String(postId.Value)
		}
	}

	return &GetPostsOutput{
		Posts:     posts,
		NextToken: nextToken,
	}, nil
}

