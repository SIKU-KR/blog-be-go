package model

import "time"

// Comment는 Partition Key로 postId, Sort Key로 commentId를 사용합니다.
type Comment struct {
	CommentID string    `json:"commentId" dynamodbav:"commentId"`
	PostID    string    `json:"postId" dynamodbav:"postId"`
	Nickname  string    `json:"nickname" dynamodbav:"nickname"`
	Content   string    `json:"content" dynamodbav:"content"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
}
