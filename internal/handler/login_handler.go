package handler

import (
	"bumsiku/internal/model"
	"bumsiku/internal/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// @Summary     관리자 로그인
// @Description 블로그 관리자 로그인 API
// @Tags        인증
// @Accept      json
// @Produce     json
// @Param       request body model.LoginRequest true "로그인 정보"
// @Success     200 {object} map[string]string "로그인 성공"
// @Failure     400 {object} ErrorResponse "잘못된 요청"
// @Failure     401 {object} ErrorResponse "로그인 실패"
// @Failure     500 {object} ErrorResponse "서버 오류"
// @Router      /login [post]
func PostLogin(c *gin.Context, logger *utils.Logger) {
	var loginVals model.LoginRequest

	if err := c.ShouldBindJSON(&loginVals); err != nil {
		contextInfo := map[string]string{
			"handler": "PostLogin",
			"step":    "요청 검증",
			"ip":      c.ClientIP(),
		}
		SendBadRequestErrorWithLogging(c, logger, "잘못된 요청 형식입니다", err, contextInfo)
		return
	}

	// 사용자 이름은 로깅하지만 암호는 로깅하지 않음
	contextInfo := map[string]string{
		"handler":  "PostLogin",
		"username": loginVals.Username,
		"ip":       c.ClientIP(),
	}

	if isValidLogin(loginVals) {
		// 로그인 실패 로깅
		logger.Warn(c.Request.Context(), "로그인 실패: 잘못된 자격 증명", contextInfo)
		SendUnauthorizedErrorWithLogging(c, logger, "로그인에 실패했습니다", nil, contextInfo)
		return
	}

	if err := activateSession(c, loginVals.Username); err != nil {
		contextInfo["step"] = "세션 활성화"
		SendInternalServerErrorWithLogging(c, logger, "세션 저장에 실패했습니다", err, contextInfo)
		return
	}

	// 로그인 성공 로깅
	logger.Info(c.Request.Context(), "관리자 로그인 성공", contextInfo)

	SendSuccess(c, http.StatusOK, map[string]string{
		"message": "로그인에 성공했습니다",
	})
}

func isValidLogin(value model.LoginRequest) bool {
	return !(value.Username == os.Getenv("ADMIN_ID") && value.Password == os.Getenv("ADMIN_PW"))
}

func activateSession(c *gin.Context, username string) error {
	session := sessions.Default(c)
	session.Set("admin", true)
	session.Set("username", username)
	session.Set("loginTime", time.Now())
	if err := session.Save(); err != nil {
		return err
	}
	return nil
}
