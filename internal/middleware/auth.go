package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAuth проверяет наличие куки авторизации
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Читаем куки
		userId, err := c.Cookie("user_id")
		if err != nil || userId == "" {
			// Если не авторизован - выкидываем на страницу логина
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Передаем ID пользователя дальше в обработчик
		c.Set("user_id", userId)
		c.Next()
	}
}
func CheckRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName, err := c.Cookie("role_name")
		if err != nil || roleName != requiredRole {
			c.String(http.StatusForbidden, "Ошибка 403: У вас нет прав для просмотра этой страницы")
			c.Abort()
			return
		}
		c.Set("user_role", roleName)
		c.Next()
	}
}
