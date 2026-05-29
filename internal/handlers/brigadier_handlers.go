package handlers

import (
	"net/http"
	"pp_sibnia/internal/database"
	"pp_sibnia/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Вспомогательная функция получения ID пользователя из куки
func getBrigadierID(c *gin.Context) uint {
	cookie, err := c.Cookie("user_id")
	if err != nil {
		return 0
	}
	id, _ := strconv.Atoi(cookie)
	return uint(id)
}

// BrigadierPage - Главная панель оператора (твоя структура данных)
func BrigadierPage(c *gin.Context) {
	brigadierID := getBrigadierID(c)

	// Получаем фокус оператора из URL
	activeShiftID, _ := strconv.Atoi(c.Query("shift_id"))
	activeConfigID, _ := strconv.Atoi(c.Query("config_id"))
	activeProtocolID, _ := strconv.Atoi(c.Query("protocol_id"))

	// 1. Загружаем смены этого бригадира (с предзагрузкой инфы об Эксперименте)
	var shifts []models.Shift
	database.DB.Preload("Experiment").Where("brigadier_id = ?", brigadierID).Order("id desc").Find(&shifts)

	// 2. Конфигурации для выбранной смены
	var configs []models.Configuration
	if activeShiftID > 0 {
		database.DB.Where("shift_id = ?", activeShiftID).Order("id asc").Find(&configs)
	}

	// 3. Протоколы для выбранной конфигурации
	var protocols []models.Protocol
	if activeConfigID > 0 {
		database.DB.Where("configuration_id = ?", activeConfigID).Order("id asc").Find(&protocols)
	}

	// 4. Шаги/Точки (ProtocolData) для выбранного протокола
	var points []models.ProtocolData
	if activeProtocolID > 0 {
		database.DB.Where("protocol_id = ?", activeProtocolID).Order("id asc").Find(&points)
	}

	// 5. Список всех существующих экспериментов (для создания смены)
	var experiments []models.Experiment
	database.DB.Order("id desc").Find(&experiments)

	c.HTML(http.StatusOK, "brigadier.html", gin.H{
		"Shifts":           shifts,
		"Configs":          configs,
		"Protocols":        protocols,
		"Points":           points,
		"Experiments":      experiments,
		"ActiveShiftID":    activeShiftID,
		"ActiveConfigID":   activeConfigID,
		"ActiveProtocolID": activeProtocolID,
	})
}

// ShiftCreate - Регистрация смены (модель Shift)
func ShiftCreate(c *gin.Context) {
	brigadierID := getBrigadierID(c)
	expID, _ := strconv.Atoi(c.PostForm("experiment_id"))
	shiftNum, _ := strconv.Atoi(c.PostForm("shift_number"))

	dateStr := c.PostForm("shift_date")
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		parsedDate = time.Now()
	}

	shift := models.Shift{
		ExperimentID:    uint(expID),
		ShiftNumber:     shiftNum,
		BrigadierID:     brigadierID,
		WorkDescription: c.PostForm("work_description"),
		ShiftDate:       parsedDate,
	}
	database.DB.Create(&shift)
	c.Redirect(http.StatusFound, "/brigadier?shift_id="+strconv.Itoa(int(shift.ID)))
}

// ConfigCreate - Создание конфигурации испытания с константами (модель Configuration)
func ConfigCreate(c *gin.Context) {
	shiftID, _ := strconv.Atoi(c.PostForm("shift_id"))
	speed, _ := strconv.ParseFloat(c.PostForm("wind_speed"), 64)
	roll, _ := strconv.ParseFloat(c.PostForm("roll_angle"), 64)
	yaw, _ := strconv.ParseFloat(c.PostForm("yaw_angle"), 64)

	config := models.Configuration{
		ShiftID:     uint(shiftID),
		Description: c.PostForm("description"),
		WindSpeed:   speed, // Скорость потока
		RollAngle:   roll,  // Крен
		YawAngle:    yaw,   // Рыскание
	}
	database.DB.Create(&config)
	c.Redirect(http.StatusFound, "/brigadier?shift_id="+strconv.Itoa(shiftID)+"&config_id="+strconv.Itoa(int(config.ID)))
}

// ProtocolCreate - Создание протокола внутри конфигурации (модель Protocol)
func ProtocolCreate(c *gin.Context) {
	shiftID := c.PostForm("shift_id")
	configID, _ := strconv.Atoi(c.PostForm("configuration_id"))

	protocol := models.Protocol{
		ConfigurationID:   uint(configID),
		ProtocolNumber:    c.PostForm("protocol_number"),
		VariableParameter: c.PostForm("variable_parameter"),
	}
	database.DB.Create(&protocol)
	c.Redirect(http.StatusFound, "/brigadier?shift_id="+shiftID+"&config_id="+strconv.Itoa(configID)+"&protocol_id="+strconv.Itoa(int(protocol.ID)))
}

// PointAdd - Пошаговое добавление экспериментальных данных (модель ProtocolData)
// PointAdd - Пошаговое добавление ВСЕХ экспериментальных данных (строго по твоей схеме БД)
func PointAdd(c *gin.Context) {
	shiftID := c.PostForm("shift_id")
	configID := c.PostForm("config_id")
	protocolID, _ := strconv.Atoi(c.PostForm("protocol_id"))

	// Парсим все физические параметры и коэффициенты аэродинамической трубы
	al, _ := strconv.ParseFloat(c.PostForm("al"), 64)
	alpha, _ := strconv.ParseFloat(c.PostForm("alpha"), 64)
	beta, _ := strconv.ParseFloat(c.PostForm("beta"), 64)
	q, _ := strconv.ParseFloat(c.PostForm("q"), 64)
	v, _ := strconv.ParseFloat(c.PostForm("v"), 64)
	pf, _ := strconv.ParseFloat(c.PostForm("pf"), 64)
	pa, _ := strconv.ParseFloat(c.PostForm("pa"), 64)
	tf, _ := strconv.ParseFloat(c.PostForm("tf"), 64)
	x, _ := strconv.ParseFloat(c.PostForm("x"), 64)
	y, _ := strconv.ParseFloat(c.PostForm("y"), 64)
	z, _ := strconv.ParseFloat(c.PostForm("z"), 64)
	mx, _ := strconv.ParseFloat(c.PostForm("mx"), 64)
	my, _ := strconv.ParseFloat(c.PostForm("my"), 64)
	mz, _ := strconv.ParseFloat(c.PostForm("mz"), 64)

	pData := models.ProtocolData{
		ProtocolID: uint(protocolID),
		PointN:     c.PostForm("point_n"), // Номер точки
		Al:         al,
		Alpha:      alpha, // Угол атаки
		Beta:       beta,  // Угол скольжения
		Q:          q,     // Скоростной напор
		V:          v,     // Скорость
		Pf:         pf,    // Давление pf
		Pa:         pa,    // Давление pa
		Tf:         tf,    // Температура
		X:          x,     // Сила X
		Y:          y,     // Сила Y
		Z:          z,     // Сила Z
		Mx:         mx,    // Момент Mx
		My:         my,    // Момент My
		Mz:         mz,    // Момент Mz
	}

	database.DB.Create(&pData)

	// Возвращаем фокус оператора на ту же страницу, смену, конфигурацию и протокол
	c.Redirect(http.StatusFound, "/brigadier?shift_id="+shiftID+"&config_id="+configID+"&protocol_id="+strconv.Itoa(protocolID))
}
