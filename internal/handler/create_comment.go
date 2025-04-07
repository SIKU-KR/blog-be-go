package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateCommentRequest는 댓글 생성 요청 구조체입니다.
type CreateCommentRequest struct {
	Nickname string `json:"nickname" binding:"required" example:"익명사용자"`   // 닉네임
	Content  string `json:"content" binding:"required" example:"좋은 글이네요!"` // 댓글 내용
}

// @Summary     댓글 등록
// @Description 특정 게시물에 새 댓글을 등록합니다
// @Tags        댓글
// @Accept      json
// @Produce     json
// @Param       postId path string true "게시물 ID"
// @Param       request body CreateCommentRequest true "댓글 정보"
// @Success     201 {object} model.Comment
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     404 {object} ErrorResponse "게시물을 찾을 수 없음"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /comments/{postId} [post]
// CreateComment는 특정 게시글에 댓글을 등록하는 핸들러입니다.
func CreateComment(
	commentRepo repository.CommentRepositoryInterface,
	postRepo repository.PostRepositoryInterface,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 요청 바디 검증
		var req CreateCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			SendBadRequestError(c, "요청 형식이 올바르지 않습니다: "+err.Error())
			return
		}

		// 2. 게시글 ID 추출
		postID := c.Param("postId")
		if postID == "" {
			SendBadRequestError(c, "게시글 ID가 필요합니다")
			return
		}

		// 3. 게시글 존재 여부 확인 (옵션)
		if postRepo != nil {
			post, err := postRepo.GetPostByID(c.Request.Context(), postID)
			if err != nil {
				SendInternalServerError(c, "게시글 확인 중 오류가 발생했습니다")
				return
			}
			if post == nil {
				SendNotFoundError(c, "존재하지 않는 게시글입니다")
				return
			}
		}

		// 4. Comment 모델 생성
		comment := &model.Comment{
			PostID:   postID,
			Nickname: req.Nickname,
			Content:  req.Content,
		}

		// 5. 댓글 저장
		createdComment, err := commentRepo.CreateComment(c.Request.Context(), comment)
		if err != nil {
			SendInternalServerError(c, "댓글 등록에 실패했습니다: "+err.Error())
			return
		}

		// 6. 성공 응답
		SendSuccess(c, http.StatusCreated, createdComment)
	}
}
