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
	Category *string
	Page     int32
	PageSize int32
}

type GetPostsOutput struct {
	Posts      []model.Post
	TotalCount int64
}

func (r *PostRepository) GetPosts(ctx context.Context, input *GetPostsInput) (*GetPostsOutput, error) {
	// 페이지네이션 계산
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = PageSize
	}

	// 카테고리 파라미터에 따라 처리 방식 결정
	if input.Category != nil && *input.Category != "" {
		// 카테고리가 있는 경우 인덱스를 사용한 Query 수행
		return r.getPostsByCategory(ctx, *input.Category, input.Page, input.PageSize)
	} else {
		// 카테고리가 없는 경우 Scan 작업으로 모든 게시글 조회
		return r.getAllPosts(ctx, input.Page, input.PageSize)
	}
}

// 특정 카테고리의 게시글을 조회하는 함수
func (r *PostRepository) getPostsByCategory(ctx context.Context, category string, page int32, pageSize int32) (*GetPostsOutput, error) {
	// 카테고리 인덱스 기반 표현식 생성
	expr, err := buildPostListExpression(category)
	if err != nil {
		return nil, err
	}

	// Content 필드를 제외한 프로젝션 표현식 생성
	projectionExp := "postId, title, createdAt, updatedAt, summary, category"

	// 총 개수 조회
	countResult, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(PostTableName),
		IndexName:                 aws.String("category-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		Select:                    types.SelectCount,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	// 페이지네이션 계산
	offset := (page - 1) * pageSize

	// 게시글 조회 쿼리
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(PostTableName),
		IndexName:                 aws.String("category-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      aws.String(projectionExp), // Content 필드 제외
		Limit:                     aws.Int32(pageSize),
		ScanIndexForward:          aws.Bool(false), // 최신순 정렬
	}

	// 오프셋 처리
	if offset > 0 {
		var lastEvaluatedKey map[string]types.AttributeValue
		for i := int32(0); i < offset/pageSize; i++ {
			tempResult, err := r.client.Query(ctx, queryInput)
			if err != nil {
				return nil, err
			}
			if tempResult.LastEvaluatedKey == nil {
				break
			}
			lastEvaluatedKey = tempResult.LastEvaluatedKey
		}
		if lastEvaluatedKey != nil {
			queryInput.ExclusiveStartKey = lastEvaluatedKey
		}
	}

	// 결과 조회
	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	// 결과 변환
	posts := make([]model.Post, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &posts)
	if err != nil {
		return nil, err
	}

	return &GetPostsOutput{
		Posts:      posts,
		TotalCount: int64(countResult.Count),
	}, nil
}

// 모든 게시글을 조회하는 함수 (카테고리 필터 없음)
func (r *PostRepository) getAllPosts(ctx context.Context, page int32, pageSize int32) (*GetPostsOutput, error) {
	// Content 필드를 제외한 프로젝션 표현식 생성
	projectionExp := "postId, title, createdAt, updatedAt, summary, category"

	// 총 개수 조회를 위한 Scan
	countResult, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(PostTableName),
		Select:    types.SelectCount,
	})
	if err != nil {
		return nil, err
	}

	// 페이지네이션 계산
	offset := (page - 1) * pageSize

	// 게시글 조회를 위한 Scan
	scanInput := &dynamodb.ScanInput{
		TableName:            aws.String(PostTableName),
		ProjectionExpression: aws.String(projectionExp), // Content 필드 제외
		Limit:                aws.Int32(pageSize),
	}

	// 오프셋 처리
	if offset > 0 {
		var lastEvaluatedKey map[string]types.AttributeValue
		for i := int32(0); i < offset/pageSize; i++ {
			tempResult, err := r.client.Scan(ctx, scanInput)
			if err != nil {
				return nil, err
			}
			if tempResult.LastEvaluatedKey == nil {
				break
			}
			lastEvaluatedKey = tempResult.LastEvaluatedKey
		}
		if lastEvaluatedKey != nil {
			scanInput.ExclusiveStartKey = lastEvaluatedKey
		}
	}

	// 결과 조회
	result, err := r.client.Scan(ctx, scanInput)
	if err != nil {
		return nil, err
	}

	// 결과 변환
	posts := make([]model.Post, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &posts)
	if err != nil {
		return nil, err
	}

	return &GetPostsOutput{
		Posts:      posts,
		TotalCount: int64(countResult.Count),
	}, nil
}

func buildPostListExpression(category string) (expression.Expression, error) {
	// 카테고리별 인덱스를 사용한 쿼리 조건 생성
	keyCondition := expression.Key("category").Equal(expression.Value(category))

	return expression.NewBuilder().
		WithKeyCondition(keyCondition).
		Build()
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
