package model

import "time"

// Category는 Primary Key로 categoryId를 사용합니다.
type Category struct {
	CategoryID string    `json:"categoryId" dynamodbav:"categoryId" example:"tech"`               // 카테고리 ID
	Name       string    `json:"name" dynamodbav:"name" example:"기술"`                             // 카테고리 이름
	CreatedAt  time.Time `json:"createdAt" dynamodbav:"createdAt" example:"2023-01-01T00:00:00Z"` // 생성 시간
}
