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

// PostTableName은 DynamoDB 게시글 테이블 이름입니다.
const PostTableName = "blog_posts"

// PageSize는 한 페이지에 표시할 게시글 수의 기본값입니다.
const PageSize = 10

// PostRepositoryInterface는 게시글 저장소의 인터페이스를 정의합니다.
type PostRepositoryInterface interface {
	GetPosts(ctx context.Context, input *GetPostsInput) (*GetPostsOutput, error)
	GetPostByID(ctx context.Context, postID string) (*domain.Post, error)
}

// PostRepository는 DynamoDB를 사용하는 게시글 저장소입니다.
type PostRepository struct {
	client *dynamodb.Client
}

// NewPostRepository는 새로운 PostRepository 인스턴스를 생성합니다.
func NewPostRepository(client *dynamodb.Client) *PostRepository {
	return &PostRepository{client: client}
}

// GetPostsInput은 게시글 목록 조회에 필요한 입력 파라미터를 정의합니다.
type GetPostsInput struct {
	Category  *string
	NextToken *string
	PageSize  *int32
}

// GetPostsOutput은 게시글 목록 조회 결과를 담는 구조체입니다.
type GetPostsOutput struct {
	Posts     []domain.Post
	NextToken *string
}

// GetPosts는 게시글 목록을 조회합니다.
func (r *PostRepository) GetPosts(ctx context.Context, input *GetPostsInput) (*GetPostsOutput, error) {
	// 기본 쿼리 설정
	keyEx := expression.Key("postId").BeginsWith("")

	// 프로젝션 표현식 추가 - 필요한 필드만 가져오기
	proj := expression.NamesList(
		expression.Name("postId"),
		expression.Name("title"),
		expression.Name("createdAt"),
		expression.Name("updatedAt"),
		expression.Name("summary"),
	)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyEx).
		WithProjection(proj).
		Build()
	if err != nil {
		return nil, err
	}

	// 페이지 크기 설정
	limit := PageSize
	if input.PageSize != nil {
		limit = int(*input.PageSize)
	}

	// 쿼리 입력 구성
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(PostTableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		Limit:                     aws.Int32(int32(limit)),
		ScanIndexForward:          aws.Bool(false), // 최신순 정렬
	}

	// 카테고리 필터링이 있는 경우
	if input.Category != nil && *input.Category != "" {
		queryInput.IndexName = aws.String("category-index")
		keyEx = expression.Key("category").Equal(expression.Value(*input.Category))
		expr, err = expression.NewBuilder().
			WithKeyCondition(keyEx).
			WithProjection(proj).
			Build()
		if err != nil {
			return nil, err
		}
		queryInput.KeyConditionExpression = expr.KeyCondition()
		queryInput.ExpressionAttributeNames = expr.Names()
		queryInput.ExpressionAttributeValues = expr.Values()
		queryInput.ProjectionExpression = expr.Projection()
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

// GetPostByID는 ID로 특정 게시글을 조회합니다.
func (r *PostRepository) GetPostByID(ctx context.Context, postID string) (*domain.Post, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(PostTableName),
		Key: map[string]types.AttributeValue{
			"postId": &types.AttributeValueMemberS{Value: postID},
		},
	})

	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	post := &domain.Post{}
	err = attributevalue.UnmarshalMap(result.Item, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}
