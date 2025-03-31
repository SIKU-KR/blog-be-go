package domain

import "time"

// Category는 Primary Key로 categoryId를 사용합니다.
type Category struct {
	CategoryID string    `json:"categoryId" dynamodbav:"categoryId"`
	Name       string    `json:"name" dynamodbav:"name"`
	CreatedAt  time.Time `json:"createdAt" dynamodbav:"createdAt"`
}
