package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"pp_sibnia/internal/database"
	"pp_sibnia/internal/handlers"
	"pp_sibnia/internal/middleware"
	"pp_sibnia/internal/models"
)

func main() {
	// 1. Загружаем .env (он лежит в той же папке, что и этот main.go)
	if err := godotenv.Load(); err != nil {
		log.Println("Внимание: файл .env не найден, берутся системные переменные")
	}

	// 2. Подключаем БД и запускаем миграции
	database.Connect()

	// 3. Настраиваем Gin
	r := gin.Default()

	// Указываем путь к HTML шаблонам
	r.LoadHTMLGlob("templates/*")

	// Маршруты авторизации
	r.GET("/", func(c *gin.Context) { c.Redirect(302, "/login") })
	r.GET("/login", handlers.LoginPage)
	r.POST("/login", handlers.LoginPost)
	r.GET("/logout", handlers.Logout)

	// Защищенная зона
	authorized := r.Group("/")
	authorized.Use(middleware.RequireAuth())
	{
		// Внутри авторизованной зоны (например, где сидят инженеры)
		authorized.GET("/reports/export", handlers.ExportExperimentXLSX)

		adminGroup := authorized.Group("/admin")
		adminGroup.Use(middleware.CheckRole("Администратор"))
		{
			adminGroup.GET("/", handlers.AdminPage)

			// Пользователи
			adminGroup.POST("/users/add", handlers.UserAdd)
			adminGroup.POST("/users/edit", handlers.UserEdit)
			adminGroup.POST("/users/delete", handlers.UserDelete)

			// Ведение НСИ (Справочники) - Добавление
			adminGroup.POST("/tunnels/add", handlers.TunnelAdd)
			adminGroup.POST("/exptypes/add", handlers.ExpTypeAdd)
			adminGroup.POST("/equipment/add", handlers.EquipmentAdd)

			// НОВОЕ: Ведение НСИ (Справочники) - Удаление
			adminGroup.POST("/tunnels/delete", handlers.TunnelDelete)
			adminGroup.POST("/exptypes/delete", handlers.ExpTypeDelete)
			adminGroup.POST("/equipment/delete", handlers.EquipmentDelete)
		}

		// Группа для Бригадира смены
		brigadierGroup := authorized.Group("/brigadier")
		brigadierGroup.Use(middleware.CheckRole("Бригадир смены"))
		{
			brigadierGroup.GET("/", handlers.BrigadierPage)
			brigadierGroup.POST("/shifts/create", handlers.ShiftCreate)
			brigadierGroup.POST("/configs/create", handlers.ConfigCreate)
			brigadierGroup.POST("/protocols/create", handlers.ProtocolCreate)
			brigadierGroup.POST("/points/add", handlers.PointAdd)
		}

		// Группа Ведущего инженера
		engineerGroup := authorized.Group("/engineer")
		engineerGroup.Use(middleware.CheckRole("Ведущий инженер"))
		{
			engineerGroup.GET("/", handlers.EngineerDashboard)
			engineerGroup.POST("/models/create", handlers.ModelLACreate)
			engineerGroup.POST("/experiments/register", handlers.ExperimentRegister)
			engineerGroup.POST("/equipment/link", handlers.ExperimentLinkEquipment)
		}

		// Группа Начальника аэродинамической трубы
		chiefGroup := authorized.Group("/tunnel_chief")
		chiefGroup.Use(middleware.CheckRole("Начальник трубы"))
		{
			chiefGroup.GET("/", handlers.EngineerDashboard)
			chiefGroup.POST("/models/create", handlers.ModelLACreate)
			chiefGroup.POST("/experiments/register", handlers.ExperimentRegister)
			chiefGroup.POST("/equipment/link", handlers.ExperimentLinkEquipment)
		}

		analystGroup := r.Group("/analyst")
		// Подключаем мидлварь проверки авторизации и роли
		analystGroup.Use(middleware.RequireAuth(), middleware.CheckRole("Аналитик"))
		{
			// Главная страница аналитика
			analystGroup.GET("", handlers.AnalystDashboard)
		}

		// Пример API добавления трубы
		authorized.POST("/api/tunnels/add", func(c *gin.Context) {
			name := c.PostForm("name")
			database.DB.Create(&models.WindTunnel{Name: name})
			c.Redirect(302, "/admin")
		})
	}

	log.Println("Сервер успешно запущен на http://localhost:8080")
	r.Run(":8080")
}
