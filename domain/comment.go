package domain

import "time"

// Partition Key: postId, Sort Key: commentId
type Comment struct {
	CommentID string    `json:"commentId" dynamodbav:"commentId"`
	PostID    string    `json:"postId" dynamodbav:"postId"`
	Nickname  string    `json:"nickname" dynamodbav:"nickname"`
	Content   string    `json:"content" dynamodbav:"content"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}