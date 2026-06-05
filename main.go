package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"pp_sibnia/internal/database"
	"pp_sibnia/internal/handlers"
	"pp_sibnia/internal/middleware"
)

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("Внимание: файл .env не найден, берутся системные переменные")
	}

	// Подключаем БД и запускаем миграции
	database.Connect()

	// Настраиваем Gin
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	// Маршруты авторизации
	r.GET("/", func(c *gin.Context) { c.Redirect(302, "/login") })
	r.GET("/login", handlers.LoginPage)
	r.POST("/login", handlers.LoginPost)
	r.GET("/logout", handlers.Logout)

	authorized := r.Group("/")
	authorized.Use(middleware.RequireAuth())
	{
		authorized.GET("/reports/export", handlers.ExportExperimentXLSX)

		adminGroup := authorized.Group("/admin")
		adminGroup.Use(middleware.CheckRole("Администратор"))
		{
			adminGroup.GET("/", handlers.AdminPage)

			adminGroup.POST("/users/add", handlers.UserAdd)
			adminGroup.POST("/users/edit", handlers.UserEdit)
			adminGroup.POST("/users/delete", handlers.UserDelete)

			adminGroup.POST("/tunnels/add", handlers.TunnelAdd)
			adminGroup.POST("/exptypes/add", handlers.ExpTypeAdd)
			adminGroup.POST("/equipment/add", handlers.EquipmentAdd)

			adminGroup.POST("/tunnels/delete", handlers.TunnelDelete)
			adminGroup.POST("/exptypes/delete", handlers.ExpTypeDelete)
			adminGroup.POST("/equipment/delete", handlers.EquipmentDelete)
		}

		brigadierGroup := authorized.Group("/brigadier")
		brigadierGroup.Use(middleware.CheckRole("Бригадир смены"))
		{
			brigadierGroup.GET("/", handlers.BrigadierPage)
			brigadierGroup.POST("/shifts/create", handlers.ShiftCreate)
			brigadierGroup.POST("/configs/create", handlers.ConfigCreate)
			brigadierGroup.POST("/protocols/create", handlers.ProtocolCreate)
			brigadierGroup.POST("/points/add", handlers.PointAdd)
		}

		engineerGroup := authorized.Group("/engineer")
		engineerGroup.Use(middleware.CheckRole("Ведущий инженер"))
		{
			engineerGroup.GET("/", handlers.EngineerDashboard)
			engineerGroup.POST("/models/create", handlers.ModelLACreate)
			engineerGroup.POST("/experiments/register", handlers.ExperimentRegister)
			engineerGroup.POST("/equipment/link", handlers.ExperimentLinkEquipment)
		}

		chiefGroup := authorized.Group("/tunnel_chief")
		chiefGroup.Use(middleware.CheckRole("Начальник трубы"))
		{
			chiefGroup.GET("/", handlers.EngineerDashboard)
			chiefGroup.POST("/models/create", handlers.ModelLACreate)
			chiefGroup.POST("/experiments/register", handlers.ExperimentRegister)
			chiefGroup.POST("/equipment/link", handlers.ExperimentLinkEquipment)
		}

		analystGroup := r.Group("/analyst")

		analystGroup.Use(middleware.RequireAuth(), middleware.CheckRole("Аналитик"))
		{

			analystGroup.GET("", handlers.AnalystDashboard)
		}

	}

	log.Println("Сервер успешно запущен на http://localhost:8080")
	r.Run(":8080")
}
