package domain

import "time"

// Partition Key: postId, Sort Key: createdAt
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
