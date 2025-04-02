package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// CreatePostRequest는 게시글 생성 요청 구조체입니다.
type CreatePostRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Summary  string `json:"summary" binding:"required"`
	Category string `json:"category" binding:"required"`
}

// CreatePost는 관리자 전용 게시글 작성 핸들러입니다.
func CreatePost(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 요청 바디 검증
		var req CreatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			SendBadRequestError(c, "요청 형식이 올바르지 않습니다: "+err.Error())
			return
		}

		// 2. nanoid 12자리 생성
		postID, err := gonanoid.New(12)
		if err != nil {
			SendInternalServerError(c, "게시글 ID 생성 실패")
			return
		}

		// 3. 현재 시간 설정
		now := time.Now()

		// 4. Post 모델 생성
		post := &model.Post{
			PostID:    postID,
			Title:     req.Title,
			Content:   req.Content,
			Summary:   req.Summary,
			Category:  req.Category,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// 5. 게시글 저장
		err = postRepo.CreatePost(c.Request.Context(), post)
		if err != nil {
			SendInternalServerError(c, "게시글 등록에 실패했습니다: "+err.Error())
			return
		}

		// 6. 성공 응답
		SendSuccess(c, http.StatusCreated, post)
	}
}
