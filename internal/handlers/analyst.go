package handlers

import (
	"net/http"
	"pp_sibnia/internal/database"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticRow описывает структуру строки сводного реестра для аналитика
type AnalyticRow struct {
	ProtocolID        uint      `json:"protocol_id"`
	ExperimentID      uint      `json:"experiment_id"` // ID эксперимента для точечной выгрузки Excel
	ProtocolNumber    string    `json:"protocol_number"`
	VariableParameter string    `json:"variable_parameter"`
	ConfigDescription string    `json:"config_description"`
	WindSpeed         float64   `json:"wind_speed"`
	ShiftNumber       int       `json:"shift_number"`
	ShiftDate         time.Time `json:"shift_date"`
	ExperimentNumber  string    `json:"experiment_number"`
	ExperimentName    string    `json:"experiment_name"`
	ContractNumber    string    `json:"contract_number"`
	PointsCount       int64     `json:"points_count"` // Количество снятых точек в протоколе
}

// AnalystDashboard обрабатывает поисковые фильтры и отображает панель аналитика
func AnalystDashboard(c *gin.Context) {
	// Получаем параметры фильтрации из URL запроса
	searchModel := c.Query("model")
	searchTunnel := c.Query("tunnel")
	searchContract := c.Query("contract")
	dateStr := c.Query("date")

	var rows []AnalyticRow

	// Формируем базовый SQL-запрос с объединением таблиц (JOIN)
	query := database.DB.Table("protocols").
		Select(`
			protocols.id as protocol_id, 
			experiments.id as experiment_id,
			protocols.protocol_number, 
			protocols.variable_parameter,
			configurations.description as config_description, 
			configurations.wind_speed,
			shifts.shift_number, 
			shifts.shift_date,
			experiments.experiment_number, 
			experiments.experiment_name,
			experiments.contract_number,
			(SELECT COUNT(*) FROM protocol_data WHERE protocol_data.protocol_id = protocols.id) as points_count
		`).
		Joins("JOIN configurations ON configurations.id = protocols.configuration_id").
		Joins("JOIN shifts ON shifts.id = configurations.shift_id").
		Joins("JOIN experiments ON experiments.id = shifts.experiment_id")

	// Применяем фильтры по мере их заполнения инженером-аналитиком
	if searchModel != "" {
		query = query.Where("experiments.experiment_name ILIKE ?", "%"+searchModel+"%")
	}
	if searchContract != "" {
		query = query.Where("experiments.contract_number ILIKE ?", "%"+searchContract+"%")
	}
	if searchTunnel != "" {
		query = query.Where("experiments.experiment_number ILIKE ?", "%"+searchTunnel+"%")
	}
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			query = query.Where("DATE(shifts.shift_date) = DATE(?)", parsedDate)
		}
	}

	// Сортируем: сначала самые свежие по дате смены и ID протокола
	query.Order("shifts.shift_date DESC, protocols.id DESC").Scan(&rows)

	// Подсчет агрегированных метрик для карточек KPI
	var totalProtocols int64 = int64(len(rows))
	var totalPoints int64 = 0
	uniqueShifts := make(map[int]bool)

	for _, row := range rows {
		totalPoints += row.PointsCount
		uniqueShifts[row.ShiftNumber] = true
	}

	// Возвращаем HTML-страницу со всеми вычисленными данными и фильтрами
	c.HTML(http.StatusOK, "analyst.html", gin.H{
		"Rows":           rows,
		"TotalProtocols": totalProtocols,
		"TotalPoints":    totalPoints,
		"TotalShifts":    len(uniqueShifts),
		"FilterModel":    searchModel,
		"FilterTunnel":   searchTunnel,
		"FilterContract": searchContract,
		"FilterDate":     dateStr,
	})
}
