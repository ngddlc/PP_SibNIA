package handlers

import (
	"net/http"
	"pp_sibnia/internal/database"
	"pp_sibnia/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func LoginPost(c *gin.Context) {
	login := c.PostForm("login")
	password := c.PostForm("password")

	var user models.User
	// 1. Ищем пользователя
	if err := database.DB.Preload("Role").Where("login = ?", login).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// 2. Проверяем пароль
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// 3. Пишем куки
	c.SetCookie("user_id", strconv.Itoa(int(user.ID)), 3600, "/", "", false, true)
	c.SetCookie("role_name", user.Role.Name, 3600, "/", "", false, true)

	// ИСПРАВЛЕНО: Вместо кучи if-else используем идиоматичный switch
	switch user.Role.Name {
	case "Администратор":
		c.Redirect(http.StatusFound, "/admin/")

	case "Бригадир смены":
		c.Redirect(http.StatusFound, "/brigadier")

	case "Ведущий инженер":
		// Сразу сделали задел на следующий шаг для инженеров
		c.Redirect(http.StatusFound, "/engineer")

	case "Начальник трубы":
		c.Redirect(http.StatusFound, "/tunnel_chief/")

	case "Аналитик":
		// Задел для аналитика
		c.Redirect(http.StatusFound, "/analyst")

	default:
		// Если роль не распознана, кидаем на главную
		c.Redirect(http.StatusFound, "/")
	}
}

func Logout(c *gin.Context) {
	// Удаляем куки
	c.SetCookie("user_id", "", -1, "/", "localhost", false, true)
	c.SetCookie("role_name", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/login")
}
