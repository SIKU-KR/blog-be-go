package repository

import (
	"context"

	"bumsiku/internal/model"

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
	GetPostByID(ctx context.Context, postID string) (*model.Post, error)
	CreatePost(ctx context.Context, post *model.Post) error
	UpdatePost(ctx context.Context, post *model.Post) error
	DeletePost(ctx context.Context, postID string) error
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
	Posts     []model.Post
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

func (r *PostRepository) GetPostByID(ctx context.Context, postID string) (*model.Post, error) {
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

func (r *PostRepository) CreatePost(ctx context.Context, post *model.Post) error {
	item, err := attributevalue.MarshalMap(post)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(PostTableName),
		Item:      item,
	})

	return err
}

func (r *PostRepository) UpdatePost(ctx context.Context, post *model.Post) error {
	// 먼저 게시글이 존재하는지 확인
	existingPost, err := r.GetPostByID(ctx, post.PostID)
	if err != nil {
		return err
	}
	if existingPost == nil {
		return &PostNotFoundError{PostID: post.PostID}
	}

	// 업데이트 표현식 생성
	update := expression.Set(expression.Name("title"), expression.Value(post.Title)).
		Set(expression.Name("content"), expression.Value(post.Content)).
		Set(expression.Name("summary"), expression.Value(post.Summary)).
		Set(expression.Name("category"), expression.Value(post.Category)).
		Set(expression.Name("updatedAt"), expression.Value(post.UpdatedAt))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(PostTableName),
		Key: map[string]types.AttributeValue{
			"postId": &types.AttributeValueMemberS{Value: post.PostID},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})

	return err
}

// PostNotFoundError는 게시글을 찾을 수 없을 때 발생하는 오류입니다.
type PostNotFoundError struct {
	PostID string
}

func (e *PostNotFoundError) Error() string {
	return "게시글을 찾을 수 없음: " + e.PostID
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
	posts := make([]model.Post, 0)
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

func unmarshallPostItem(item map[string]types.AttributeValue) (*model.Post, error) {
	if item == nil {
		return nil, nil
	}

	post := &model.Post{}
	err := attributevalue.UnmarshalMap(item, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *PostRepository) DeletePost(ctx context.Context, postID string) error {
	// 먼저 게시글이 존재하는지 확인
	existingPost, err := r.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}
	if existingPost == nil {
		return &PostNotFoundError{PostID: postID}
	}

	// 게시글 삭제
	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(PostTableName),
		Key: map[string]types.AttributeValue{
			"postId": &types.AttributeValueMemberS{Value: postID},
		},
	})

	return err
}
