package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/utils"
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// UploadImage는 이미지를 업로드하고 S3에 저장합니다.
// @Summary     이미지 업로드
// @Description 블로그에 표시할 이미지를 업로드합니다 (관리자 전용)
// @Tags        이미지
// @Accept      multipart/form-data
// @Produce     json
// @Security    AdminAuth
// @Param       image formData file true "이미지 파일"
// @Success     200 {object} model.UploadImageResponse "업로드 성공"
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     401 {object} ErrorResponse "인증되지 않은 요청"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /admin/images [post]
func UploadImage(s3Client *s3.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 멀티파트 폼 파일 가져오기
		file, err := c.FormFile("image")
		if err != nil {
			SendBadRequestError(c, "이미지 파일을 찾을 수 없습니다")
			return
		}

		// 파일 크기 제한 체크 (10MB)
		if file.Size > 10*1024*1024 {
			SendBadRequestError(c, "파일 크기는 10MB 이하여야 합니다")
			return
		}

		// 이미지 처리 및 S3 업로드
		ctx := context.Background()
		webpBytes, fileName, s3URL, err := utils.ProcessImage(ctx, s3Client, file)
		if err != nil {
			SendInternalServerError(c, "이미지 처리 중 오류가 발생했습니다: "+err.Error())
			return
		}

		// 응답 생성
		response := model.UploadImageResponse{
			URL:       s3URL,
			FileName:  fileName,
			Size:      int64(len(webpBytes)),
			MimeType:  "image/webp",
			Timestamp: time.Now().Unix(),
		}

		c.JSON(http.StatusOK, response)
	}
}
