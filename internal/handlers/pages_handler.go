package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Отдача страницы логина
func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}
