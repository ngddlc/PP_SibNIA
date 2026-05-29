package handlers

import (
	"net/http"
	"pp_sibnia/internal/database"
	"pp_sibnia/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ExperimentReport - Расширенная структура эксперимента для вывода приборов на фронтенд
type ExperimentReport struct {
	models.Experiment
	EquipmentList []string
}

// EngineerDashboard - Панель управления для Ведущего инженера и Начальника трубы
func EngineerDashboard(c *gin.Context) {
	// 1. Безопасно извлекаем роль (избегаем паники interface conversion)
	var userRoleStr string

	if val, exists := c.Get("user_role"); exists && val != nil {
		if s, ok := val.(string); ok {
			userRoleStr = s
		}
	}

	// Подстраховка: если в контексте пусто, берем роль напрямую из куки браузера
	if userRoleStr == "" {
		if cookieRole, err := c.Cookie("role_name"); err == nil && cookieRole != "" {
			userRoleStr = cookieRole
		} else {
			// Дефолтное значение на случай непредвиденного сброса сессии
			userRoleStr = "Ведущий инженер"
		}
	}

	roleTitle := "Панель управления: " + userRoleStr

	// 2. Получаем нормативно-справочную информацию (НСИ) для списков
	var tunnels []models.WindTunnel
	database.DB.Order("name asc").Find(&tunnels)

	var aircraftModels []models.ModelLA
	database.DB.Order("model_number asc").Find(&aircraftModels)

	var equipment []models.Equipment
	database.DB.Order("name asc").Find(&equipment)

	// 3. Выбираем персонал для селекторов назначения ответственных
	var chiefs []models.User
	database.DB.Joins("Role").Where("\"Role\".name = ?", "Начальник трубы").Find(&chiefs)

	var engineers []models.User
	database.DB.Joins("Role").Where("\"Role\".name = ?", "Ведущий инженер").Find(&engineers)

	// 4. Загружаем список всех экспериментов со всеми внешними связями
	var experiments []models.Experiment
	database.DB.Preload("WindTunnel").Preload("ModelLA").Preload("TunnelChief").Preload("LeadEngineer").Order("id desc").Find(&experiments)

	// 5. Собираем состав используемого оборудования для каждого эксперимента через Pluck
	var reports []ExperimentReport
	for _, exp := range experiments {
		var equipNames []string

		database.DB.Table("experiment_equipment").
			Joins("JOIN equipment ON equipment.id = experiment_equipment.equipment_id").
			Where("experiment_equipment.experiment_id = ?", exp.ID).
			Pluck("equipment.name", &equipNames)

		reports = append(reports, ExperimentReport{
			Experiment:    exp,
			EquipmentList: equipNames,
		})
	}

	// Отправляем собранный пакет данных на единый универсальный HTML-шаблон
	c.HTML(http.StatusOK, "engineer.html", gin.H{
		"RoleTitle":      roleTitle,
		"UserRole":       userRoleStr,
		"Tunnels":        tunnels,
		"AircraftModels": aircraftModels,
		"Equipment":      equipment,
		"Chiefs":         chiefs,
		"Engineers":      engineers,
		"Reports":        reports,
	})
}

// ModelLACreate - Утверждение кодового названия и номера модели ЛА в БД
func ModelLACreate(c *gin.Context) {
	redirectPath := c.PostForm("redirect_path")

	model := models.ModelLA{
		ModelNumber: c.PostForm("model_number"),
		CodeName:    c.PostForm("code_name"),
	}

	database.DB.Create(&model)
	c.Redirect(http.StatusFound, redirectPath)
}

// ExperimentRegister - Инициация эксперимента (привязка к договору, трубе и контроль сроков)
func ExperimentRegister(c *gin.Context) {
	redirectPath := c.PostForm("redirect_path")

	tunnelID, _ := strconv.Atoi(c.PostForm("wind_tunnel_id"))
	modelID, _ := strconv.Atoi(c.PostForm("model_id"))
	chiefID, _ := strconv.Atoi(c.PostForm("tunnel_chief_id"))
	engineerID, _ := strconv.Atoi(c.PostForm("lead_engineer_id"))

	// Парсинг дат временных рамок эксперимента
	startParsed, _ := time.Parse("2006-01-02", c.PostForm("start_date"))

	var endParsed *time.Time
	if endDateStr := c.PostForm("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endParsed = &t
		}
	}

	exp := models.Experiment{
		ExperimentNumber: c.PostForm("experiment_number"),
		ExperimentName:   c.PostForm("experiment_name"),
		ContractNumber:   c.PostForm("contract_number"),
		WindTunnelID:     uint(tunnelID),
		ModelID:          uint(modelID),
		StartDate:        startParsed,
		EndDate:          endParsed,
		TunnelChiefID:    uint(chiefID),
		LeadEngineerID:   uint(engineerID),
	}

	database.DB.Create(&exp)
	c.Redirect(http.StatusFound, redirectPath)
}

// ExperimentLinkEquipment - Определение состава используемого оборудования (Many-to-Many)
func ExperimentLinkEquipment(c *gin.Context) {
	redirectPath := c.PostForm("redirect_path")
	expID, _ := strconv.Atoi(c.PostForm("experiment_id"))
	equipID, _ := strconv.Atoi(c.PostForm("equipment_id"))

	link := models.ExperimentEquipment{
		ExperimentID: uint(expID),
		EquipmentID:  uint(equipID),
	}

	// FirstOrCreate страхует от дублирования записей при повторных кликах по одной кнопке
	database.DB.FirstOrCreate(&link, models.ExperimentEquipment{
		ExperimentID: uint(expID),
		EquipmentID:  uint(equipID),
	})

	c.Redirect(http.StatusFound, redirectPath)
}
