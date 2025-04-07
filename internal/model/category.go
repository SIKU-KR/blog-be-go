package model

import "time"

// Category는 Primary Key로 Category를 사용합니다.
type Category struct {
	Category  string    `json:"category" dynamodbav:"category" example:"tech"`                   // 카테고리 (기본키)
	Order     int       `json:"order" dynamodbav:"order" example:"1"`                            // 카테고리 순서
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt" example:"2023-01-01T00:00:00Z"` // 생성 시간
}
