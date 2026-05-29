package handlers

import (
	"net/http"
	"pp_sibnia/internal/database"
	"pp_sibnia/internal/models"
	"strconv"
	"unicode"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Вспомогательная функция для проверки сложности пароля
func isPasswordValid(password string) (bool, string) {
	if len(password) < 6 {
		return false, "Пароль должен быть не менее 6 символов длиной."
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return false, "Пароль должен содержать как минимум одну букву и одну цифру."
	}
	return true, ""
}

// AdminPage - рендер главной панели со всеми справочниками + вывод ошибок
func AdminPage(c *gin.Context) {
	// Читаем ошибку из URL, если она была передана при редиректе
	errMsg := c.Query("error")

	var users []models.User
	database.DB.Preload("Role").Order("id asc").Find(&users)

	var roles []models.Role
	database.DB.Find(&roles)

	var tunnels []models.WindTunnel
	database.DB.Order("id asc").Find(&tunnels)

	var expTypes []models.ExperimentType
	database.DB.Order("id asc").Find(&expTypes)

	var equipment []models.Equipment
	database.DB.Order("id asc").Find(&equipment)

	c.HTML(http.StatusOK, "admin.html", gin.H{
		"Users":     users,
		"Roles":     roles,
		"Tunnels":   tunnels,
		"ExpTypes":  expTypes,
		"Equipment": equipment,
		"Error":     errMsg, // Передаем ошибку в HTML шаблон
	})
}

// UserAdd - Создание пользователя с валидацией пароля
func UserAdd(c *gin.Context) {
	password := c.PostForm("password")

	// БЭКЕНД-ВАЛИДАЦИЯ ПАРОЛЯ
	if ok, msg := isPasswordValid(password); !ok {
		c.Redirect(http.StatusFound, "/admin/?error="+msg)
		return
	}

	roleID, _ := strconv.Atoi(c.PostForm("role_id"))
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Login:          c.PostForm("login"),
		PasswordHash:   string(hashedPassword),
		LastName:       c.PostForm("last_name"),
		FirstName:      c.PostForm("first_name"),
		MiddleName:     c.PostForm("middle_name"),
		RoleID:         uint(roleID),
		ContactNumber:  c.PostForm("contact_number"),
		ContractNumber: c.PostForm("contract_number"),
	}

	database.DB.Create(&user)
	c.Redirect(http.StatusFound, "/admin/")
}

// UserEdit - Редактирование пользователя с валидацией нового пароля
func UserEdit(c *gin.Context) {
	userID := c.PostForm("user_id")
	var user models.User

	if err := database.DB.First(&user, userID).Error; err == nil {
		user.Login = c.PostForm("login")
		user.LastName = c.PostForm("last_name")
		user.FirstName = c.PostForm("first_name")
		user.MiddleName = c.PostForm("middle_name")
		user.ContactNumber = c.PostForm("contact_number")
		user.ContractNumber = c.PostForm("contract_number")

		roleID, _ := strconv.Atoi(c.PostForm("role_id"))
		user.RoleID = uint(roleID)

		newPassword := c.PostForm("password")
		if newPassword != "" {
			// БЭКЕНД-ВАЛИДАЦИЯ ПАРОЛЯ ПРИ ИЗМЕНЕНИИ
			if ok, msg := isPasswordValid(newPassword); !ok {
				c.Redirect(http.StatusFound, "/admin/?error="+msg)
				return
			}
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			user.PasswordHash = string(hashedPassword)
		}

		database.DB.Save(&user)
	}
	c.Redirect(http.StatusFound, "/admin/")
}

// UserDelete - Удаление пользователя
func UserDelete(c *gin.Context) {
	userID := c.PostForm("user_id")
	database.DB.Delete(&models.User{}, userID)
	c.Redirect(http.StatusFound, "/admin/")
}

// TunnelAdd - Добавление аэродинамической трубы
func TunnelAdd(c *gin.Context) {
	name := c.PostForm("name")
	if name != "" {
		database.DB.Create(&models.WindTunnel{Name: name})
	}
	c.Redirect(http.StatusFound, "/admin/")
}

// TunnelDelete - НОВОЕ: Удаление трубы
func TunnelDelete(c *gin.Context) {
	id := c.PostForm("id")
	database.DB.Delete(&models.WindTunnel{}, id)
	c.Redirect(http.StatusFound, "/admin/")
}

// ExpTypeAdd - Добавление нового вида эксперимента
func ExpTypeAdd(c *gin.Context) {
	name := c.PostForm("name")
	if name != "" {
		database.DB.Create(&models.ExperimentType{Name: name})
	}
	c.Redirect(http.StatusFound, "/admin/")
}

// ExpTypeDelete - НОВОЕ: Удаление вида эксперимента
func ExpTypeDelete(c *gin.Context) {
	id := c.PostForm("id")
	database.DB.Delete(&models.ExperimentType{}, id)
	c.Redirect(http.StatusFound, "/admin/")
}

// EquipmentAdd - Добавление измерительного прибора
func EquipmentAdd(c *gin.Context) {
	name := c.PostForm("name")
	codeName := c.PostForm("code_name")
	if name != "" && codeName != "" {
		database.DB.Create(&models.Equipment{Name: name, CodeName: codeName})
	}
	c.Redirect(http.StatusFound, "/admin/")
}

// EquipmentDelete - НОВОЕ: Удаление измерительного прибора
func EquipmentDelete(c *gin.Context) {
	id := c.PostForm("id")
	database.DB.Delete(&models.Equipment{}, id)
	c.Redirect(http.StatusFound, "/admin/")
}
