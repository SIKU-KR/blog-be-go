package repository

import (
	"context"
	"sort"

	"bumsiku/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const CategoryTableName = "blog_categories"

type CategoryRepositoryInterface interface {
	GetCategories(ctx context.Context) ([]model.Category, error)
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
