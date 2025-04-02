package handler

import (
	"bumsiku/internal/model"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func PostLogin(c *gin.Context) {
	var loginVals model.LoginRequest

	if err := c.ShouldBindJSON(&loginVals); err != nil {
		SendBadRequestError(c, "잘못된 요청 형식입니다")
		return
	}

	if isValidLogin(loginVals) {
		SendUnauthorizedError(c, "로그인에 실패했습니다")
		return
	}

	if err := activateSession(c, loginVals.Username); err != nil {
		SendInternalServerError(c, "세션 저장에 실패했습니다")
		return
	}

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
