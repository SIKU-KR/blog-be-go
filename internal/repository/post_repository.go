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

const PostTableName = "blog_posts"
const PageSize = 10

type PostRepositoryInterface interface {
	GetPosts(ctx context.Context, input *GetPostsInput) (*GetPostsOutput, error)
	GetPostByID(ctx context.Context, postID string) (*domain.Post, error)
}

type PostRepository struct {
	client *dynamodb.Client
}

func NewPostRepository(client *dynamodb.Client) *PostRepository {
	return &PostRepository{client: client}
}

type GetPostsInput struct {
	Category  *string
	NextToken *string
	PageSize  *int32
}

type GetPostsOutput struct {
	Posts     []domain.Post
	NextToken *string
}

func (r *PostRepository) GetPosts(ctx context.Context, input *GetPostsInput) (*GetPostsOutput, error) {
	queryInput, err := r.buildPostsQueryInput(input)
	if err != nil {
		return nil, err
	}

	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	return r.processPostsQueryResult(result)
}

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

	return unmarshallPostItem(result.Item)
}

func (r *PostRepository) buildPostsQueryInput(input *GetPostsInput) (*dynamodb.QueryInput, error) {
	expr, err := buildPostListExpression("")
	if err != nil {
		return nil, err
	}

	limit := resolvePageSize(input.PageSize)

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(PostTableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		Limit:                     aws.Int32(limit),
		ScanIndexForward:          aws.Bool(false), // 최신순 정렬
	}

	queryInput = r.applyFilterByCategory(queryInput, input.Category, expr)

	if input.NextToken != nil {
		queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
			"postId": &types.AttributeValueMemberS{Value: *input.NextToken},
		}
	}

	return queryInput, nil
}

func (r *PostRepository) applyFilterByCategory(queryInput *dynamodb.QueryInput, category *string, expr expression.Expression) *dynamodb.QueryInput {
	if category == nil || *category == "" {
		return queryInput
	}

	newExpr, err := buildPostListExpression(*category)
	if err != nil {
		return queryInput // 오류 발생 시 원래 쿼리 반환
	}

	queryInput.IndexName = aws.String("category-index")
	queryInput.KeyConditionExpression = newExpr.KeyCondition()
	queryInput.ExpressionAttributeNames = newExpr.Names()
	queryInput.ExpressionAttributeValues = newExpr.Values()
	queryInput.ProjectionExpression = newExpr.Projection()

	return queryInput
}

func (r *PostRepository) processPostsQueryResult(result *dynamodb.QueryOutput) (*GetPostsOutput, error) {
	// 결과 변환
	posts := make([]domain.Post, 0)
	err := attributevalue.UnmarshalListOfMaps(result.Items, &posts)
	if err != nil {
		return nil, err
	}

	// 다음 페이지 토큰 추출
	nextToken := extractNextToken(result.LastEvaluatedKey)

	return &GetPostsOutput{
		Posts:     posts,
		NextToken: nextToken,
	}, nil
}

func extractNextToken(lastEvaluatedKey map[string]types.AttributeValue) *string {
	if lastEvaluatedKey == nil {
		return nil
	}

	if postId, ok := lastEvaluatedKey["postId"].(*types.AttributeValueMemberS); ok {
		return aws.String(postId.Value)
	}

	return nil
}

func buildPostListExpression(category string) (expression.Expression, error) {
	var keyCondition expression.KeyConditionBuilder

	if category == "" {
		keyCondition = expression.Key("postId").BeginsWith("")
	} else {
		keyCondition = expression.Key("category").Equal(expression.Value(category))
	}

	projection := buildPostListProjection()

	return expression.NewBuilder().
		WithKeyCondition(keyCondition).
		WithProjection(projection).
		Build()
}

func buildPostListProjection() expression.ProjectionBuilder {
	return expression.NamesList(
		expression.Name("postId"),
		expression.Name("title"),
		expression.Name("createdAt"),
		expression.Name("updatedAt"),
		expression.Name("summary"),
	)
}

func resolvePageSize(pageSize *int32) int32 {
	if pageSize == nil {
		return PageSize
	}
	return *pageSize
}

func unmarshallPostItem(item map[string]types.AttributeValue) (*domain.Post, error) {
	if item == nil {
		return nil, nil
	}

	post := &domain.Post{}
	err := attributevalue.UnmarshalMap(item, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}
