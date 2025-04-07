package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"bumsiku/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdatePostRequest는 게시글 수정 요청 구조체입니다.
type UpdatePostRequest struct {
	Title    string `json:"title" binding:"required" example:"수정된 블로그 게시물"`        // 게시물 제목
	Content  string `json:"content" binding:"required" example:"수정된 게시물 본문 내용..."` // 게시물 내용
	Summary  string `json:"summary" binding:"required" example:"수정된 게시물 요약..."`    // 게시물 요약
	Category string `json:"category" binding:"required" example:"technology"`      // 카테고리
}

// @Summary     게시물 수정
// @Description 기존 블로그 게시물을 수정합니다 (관리자 전용)
// @Tags        게시물
// @Accept      json
// @Produce     json
// @Security    AdminAuth
// @Param       id path string true "게시물 ID"
// @Param       request body UpdatePostRequest true "수정할 게시물 정보"
// @Success     200 {object} model.Post
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     401 {object} ErrorResponse "인증 실패"
// @Failure     404 {object} ErrorResponse "게시물을 찾을 수 없음"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /admin/posts/{id} [put]
// UpdatePost는 관리자 전용 게시글 수정 핸들러입니다.
func UpdatePost(postRepo repository.PostRepositoryInterface, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 경로 파라미터 확인
		postID := c.Param("id")
		if postID == "" {
			contextInfo := map[string]string{
				"handler": "UpdatePost",
				"step":    "경로 파라미터 확인",
			}
			SendBadRequestErrorWithLogging(c, logger, "게시글 ID가 필요합니다", nil, contextInfo)
			return
		}

		// 2. 요청 바디 검증
		var req UpdatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			contextInfo := map[string]string{
				"handler": "UpdatePost",
				"step":    "요청 검증",
				"postID":  postID,
			}
			SendBadRequestErrorWithLogging(c, logger, "요청 형식이 올바르지 않습니다", err, contextInfo)
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
			contextInfo := map[string]string{
				"handler":  "UpdatePost",
				"step":     "게시글 업데이트",
				"postID":   postID,
				"category": req.Category,
				"title":    req.Title,
			}

			// PostNotFoundError 확인
			if _, ok := err.(*repository.PostNotFoundError); ok {
				SendNotFoundErrorWithLogging(c, logger, "게시글을 찾을 수 없습니다", err, contextInfo)
				return
			}

			SendInternalServerErrorWithLogging(c, logger, "게시글 수정에 실패했습니다", err, contextInfo)
			return
		}

		// 6. 수정된 게시글 조회
		updatedPost, err := postRepo.GetPostByID(c.Request.Context(), postID)
		if err != nil {
			contextInfo := map[string]string{
				"handler": "UpdatePost",
				"step":    "수정된 게시글 조회",
				"postID":  postID,
			}
			SendInternalServerErrorWithLogging(c, logger, "수정된 게시글 조회에 실패했습니다", err, contextInfo)
			return
		}

		// 로그 남기기 - 성공 케이스
		logger.Info(c.Request.Context(), "게시글이 성공적으로 수정되었습니다", map[string]string{
			"handler":   "UpdatePost",
			"postID":    postID,
			"category":  req.Category,
			"title":     req.Title,
			"updatedAt": now.Format(time.RFC3339),
		})

		// 7. 성공 응답
		SendSuccess(c, http.StatusOK, updatedPost)
	}
}
