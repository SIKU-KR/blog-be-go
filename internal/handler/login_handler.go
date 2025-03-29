package handler

import (
	"bumsiku/domain"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

func PostLogin(c *gin.Context) {
	var loginVals domain.LoginRequest

	if err := c.ShouldBindJSON(&loginVals); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	if validateLogin(loginVals) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to login"})
		return
	}

	if err := activateSession(c, loginVals.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login Successful"})
}

func validateLogin(value domain.LoginRequest) bool {
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
