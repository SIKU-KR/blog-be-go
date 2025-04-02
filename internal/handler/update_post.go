package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdatePostRequest는 게시글 수정 요청 구조체입니다.
type UpdatePostRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Summary  string `json:"summary" binding:"required"`
	Category string `json:"category" binding:"required"`
}

// UpdatePost는 관리자 전용 게시글 수정 핸들러입니다.
func UpdatePost(postRepo repository.PostRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 경로 파라미터 확인
		postID := c.Param("id")
		if postID == "" {
			SendBadRequestError(c, "게시글 ID가 필요합니다")
			return
		}

		// 2. 요청 바디 검증
		var req UpdatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			SendBadRequestError(c, "요청 형식이 올바르지 않습니다: "+err.Error())
			return
		}

		// 3. 업데이트 시간 설정
		now := time.Now()

		// 4. Post 모델 생성
		post := &model.Post{
			PostID:    postID,
			Title:     req.Title,
			Content:   req.Content,
			Summary:   req.Summary,
			Category:  req.Category,
			UpdatedAt: now,
		}

		// 5. 게시글 업데이트
		err := postRepo.UpdatePost(c.Request.Context(), post)
		if err != nil {
			// PostNotFoundError 확인
			if _, ok := err.(*repository.PostNotFoundError); ok {
				SendNotFoundError(c, err.Error())
				return
			}

			SendInternalServerError(c, "게시글 수정에 실패했습니다: "+err.Error())
			return
		}

		// 6. 수정된 게시글 조회
		updatedPost, err := postRepo.GetPostByID(c.Request.Context(), postID)
		if err != nil {
			SendInternalServerError(c, "수정된 게시글 조회에 실패했습니다: "+err.Error())
			return
		}

		// 7. 성공 응답
		SendSuccess(c, http.StatusOK, updatedPost)
	}
}
