package domain

import "time"

// Post는 블로그 게시물 정보를 담는 구조체입니다. Partition Key로 postId, Sort Key로 createdAt을 사용합니다.
// GSI: categoryId, Sort Key: createdAt
type Post struct {
	PostID    string    `json:"postId" dynamodbav:"postId"`
	Title     string    `json:"title" dynamodbav:"title"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
	Content   string    `json:"content" dynamodbav:"content"`
	Summary   string    `json:"summary" dynamodbav:"summary"`
	Category  string    `json:"category" dynamodbav:"category"`
}
