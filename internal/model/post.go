package model

import "time"

// Post는 블로그 게시물 정보를 담는 구조체입니다. Partition Key로 postId, Sort Key로 createdAt을 사용합니다.
// GSI: categoryId, Sort Key: createdAt
type Post struct {
	PostID    string    `json:"postId" dynamodbav:"postId" example:"post-123"`                   // 게시물 ID
	Title     string    `json:"title" dynamodbav:"title" example:"블로그 제목"`                       // 게시물 제목
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt" example:"2023-01-01T00:00:00Z"` // 생성 시간
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"updatedAt" example:"2023-01-01T00:00:00Z"` // 수정 시간
	Content   string    `json:"content" dynamodbav:"content" example:"게시물 본문 내용..."`             // 게시물 내용
	Summary   string    `json:"summary" dynamodbav:"summary" example:"게시물 요약..."`                // 게시물 요약
	Category  string    `json:"category" dynamodbav:"category" example:"technology"`             // 카테고리
}
