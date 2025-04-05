package repository

import (
	"context"
	"sort"
	"time"

	"bumsiku/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const CategoryTableName = "blog_categories"

type CategoryRepositoryInterface interface {
	GetCategories(ctx context.Context) ([]model.Category, error)
	UpsertCategory(ctx context.Context, category model.Category) error
}

type CategoryRepository struct {
	client *dynamodb.Client
}

func NewCategoryRepository(client *dynamodb.Client) *CategoryRepository {
	return &CategoryRepository{client: client}
}

func (r *CategoryRepository) GetCategories(ctx context.Context) ([]model.Category, error) {
	// 모든 카테고리 가져오기
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(CategoryTableName),
	})
	if err != nil {
		return nil, err
	}

	// 결과 변환
	categories := make([]model.Category, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &categories)
	if err != nil {
		return nil, err
	}

	// Order 필드 기준으로 정렬
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Order < categories[j].Order
	})

	return categories, nil
}

// UpsertCategory는 카테고리를 생성하거나 업데이트합니다
func (r *CategoryRepository) UpsertCategory(ctx context.Context, category model.Category) error {
	// CreatedAt이 설정되지 않은 경우에만 현재 시간으로 설정
	if category.CreatedAt.IsZero() {
		category.CreatedAt = time.Now()
	}

	// DynamoDB에 항목 저장
	item, err := attributevalue.MarshalMap(category)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(CategoryTableName),
		Item:      item,
	})

	return err
}
