package domain

import "time"

// Primary Key로 categoryId를 사용합니다.
type Category struct {
	CategoryID string    `json:"categoryId" dynamodbav:"categoryId"`
	Name       string    `json:"name" dynamodbav:"name"`
	CreatedAt  time.Time `json:"createdAt" dynamodbav:"createdAt"`
}
