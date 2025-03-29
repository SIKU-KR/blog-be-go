package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

func PostLogin(c *gin.Context) {
	var loginVals struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginVals); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	if loginVals.Username != os.Getenv("ADMIN_ID") || loginVals.Password != os.Getenv("ADMIN_PW") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to login"})
		return
	}

	session := sessions.Default(c)
	session.Set("admin", true)
	session.Set("username", loginVals.Username)
	session.Set("loginTime", time.Now())
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login Successful"})
}
