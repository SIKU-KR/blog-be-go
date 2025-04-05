package model

import "time"

// Comment는 Partition Key로 postId, Sort Key로 commentId를 사용합니다.
type Comment struct {
	CommentID string    `json:"commentId" dynamodbav:"commentId" example:"comment-123"`          // 댓글 ID
	PostID    string    `json:"postId" dynamodbav:"postId" example:"post-123"`                   // 게시물 ID
	Nickname  string    `json:"nickname" dynamodbav:"nickname" example:"익명사용자"`                  // 닉네임
	Content   string    `json:"content" dynamodbav:"content" example:"댓글 내용입니다."`                // 댓글 내용
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt" example:"2023-01-01T00:00:00Z"` // 생성 시간
}
