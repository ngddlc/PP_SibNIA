package handlers

import (
	"fmt"
	"net/http"
	"pp_sibnia/internal/database"
	"pp_sibnia/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ExportExperimentXLSX - Генерация отчта
func ExportExperimentXLSX(c *gin.Context) {
	expID, _ := strconv.Atoi(c.Query("experiment_id"))

	// Выкачиваем полную структуру эксперимента из БД
	var exp models.Experiment
	if err := database.DB.Preload("WindTunnel").Preload("ModelLA").Preload("TunnelChief").Preload("LeadEngineer").First(&exp, expID).Error; err != nil {
		c.String(http.StatusNotFound, "Эксперимент не найден")
		return
	}

	// Получаем состав оборудования
	var equipNames []string
	database.DB.Table("experiment_equipment").
		Joins("JOIN equipment ON equipment.id = experiment_equipment.equipment_id").
		Where("experiment_equipment.experiment_id = ?", exp.ID).
		Pluck("equipment.name", &equipNames)

	// Инициализируем Excel-файл
	f := excelize.NewFile()
	defer f.Close()

	sheet1 := "Паспорт эксперимента"
	f.SetSheetName("Sheet1", sheet1)

	f.SetCellValue(sheet1, "A1", "СВОДНЫЙ НАУЧНО-ТЕХНИЧЕСКИЙ ОТЧЕТ АЭРОДИНАМИЧЕСКИХ ИСПЫТАНИЙ")
	f.SetCellValue(sheet1, "A3", "Шифр программы:")
	f.SetCellValue(sheet1, "B3", exp.ExperimentNumber)
	f.SetCellValue(sheet1, "A4", "Наименование работ:")
	f.SetCellValue(sheet1, "B4", exp.ExperimentName)
	f.SetCellValue(sheet1, "A5", "Номер договора:")
	f.SetCellValue(sheet1, "B5", exp.ContractNumber)
	f.SetCellValue(sheet1, "A6", "Аэродинамическая труба:")
	f.SetCellValue(sheet1, "B6", exp.WindTunnel.Name)
	f.SetCellValue(sheet1, "A7", "Объект испытаний (Модель):")
	f.SetCellValue(sheet1, "B7", fmt.Sprintf("%s (%s)", exp.ModelLA.ModelNumber, exp.ModelLA.CodeName))

	f.SetCellValue(sheet1, "A9", "Плановые сроки:")
	f.SetCellValue(sheet1, "B9", fmt.Sprintf("с %s по %s", exp.StartDate.Format("02.01.2006"), func() string {
		if exp.EndDate != nil {
			return exp.EndDate.Format("02.01.2006")
		}
		return "н.в."
	}()))

	f.SetCellValue(sheet1, "A10", "Ответственные лица:")
	f.SetCellValue(sheet1, "B10", fmt.Sprintf("Нач. АDТ: %s, Вед. инженер: %s", exp.TunnelChief.LastName, exp.LeadEngineer.LastName))

	// Выводим состав оборудования
	f.SetCellValue(sheet1, "A12", "ИСПОЛЬЗУЕМЫЙ ИЗМЕРИТЕЛЬНЫЙ КОМПЛЕКС:")
	for i, eq := range equipNames {
		f.SetCellValue(sheet1, "B"+strconv.Itoa(12+i), fmt.Sprintf("- %s", eq))
	}

	f.SetColWidth(sheet1, "A", "A", 30)
	f.SetColWidth(sheet1, "B", "B", 60)

	sheet2 := "Данные продувок"
	f.NewSheet(sheet2)

	headers := []string{"Номер протокола", "Точка", "al", "Alpha (α)", "Beta (β)", "Q (напор)", "V (скорость)", "Pf", "Pa", "Tf", "Сила X", "Сила Y", "Сила Z", "Момент Mx", "Момент My", "Момент Mz"}
	for colNum, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colNum+1, 1)
		f.SetCellValue(sheet2, cell, header)
	}

	// Выбираем все точки, принадлежащие протоколам этого эксперимента
	var points []struct {
		ProtocolNumber string
		models.ProtocolData
	}

	database.DB.Table("protocol_data").
		Select("protocols.protocol_number, protocol_data.*").
		Joins("JOIN protocols ON protocols.id = protocol_data.protocol_id").
		Joins("JOIN configurations ON configurations.id = protocols.configuration_id").
		Joins("JOIN shifts ON shifts.id = configurations.shift_id").
		Where("shifts.experiment_id = ?", exp.ID).
		Order("protocols.id ASC, protocol_data.id ASC").
		Scan(&points)

	// Заполняем массив данных
	rowIdx := 2
	for _, p := range points {
		f.SetCellValue(sheet2, "A"+strconv.Itoa(rowIdx), p.ProtocolNumber)
		f.SetCellValue(sheet2, "B"+strconv.Itoa(rowIdx), p.PointN)
		f.SetCellValue(sheet2, "C"+strconv.Itoa(rowIdx), p.Al)
		f.SetCellValue(sheet2, "D"+strconv.Itoa(rowIdx), p.Alpha)
		f.SetCellValue(sheet2, "E"+strconv.Itoa(rowIdx), p.Beta)
		f.SetCellValue(sheet2, "F"+strconv.Itoa(rowIdx), p.Q)
		f.SetCellValue(sheet2, "G"+strconv.Itoa(rowIdx), p.V)
		f.SetCellValue(sheet2, "H"+strconv.Itoa(rowIdx), p.Pf)
		f.SetCellValue(sheet2, "I"+strconv.Itoa(rowIdx), p.Pa)
		f.SetCellValue(sheet2, "J"+strconv.Itoa(rowIdx), p.Tf)
		f.SetCellValue(sheet2, "K"+strconv.Itoa(rowIdx), p.X)
		f.SetCellValue(sheet2, "L"+strconv.Itoa(rowIdx), p.Y)
		f.SetCellValue(sheet2, "M"+strconv.Itoa(rowIdx), p.Z)
		f.SetCellValue(sheet2, "N"+strconv.Itoa(rowIdx), p.Mx)
		f.SetCellValue(sheet2, "O"+strconv.Itoa(rowIdx), p.My)
		f.SetCellValue(sheet2, "P"+strconv.Itoa(rowIdx), p.Mz)
		rowIdx++
	}

	for i := 1; i <= 16; i++ {
		cell, _ := excelize.CoordinatesToCellName(i, 1)
		f.SetColWidth(sheet2, cell[:1], cell[:1], 15)
	}

	// передача файла в браузер
	fileName := fmt.Sprintf("SibNIA_Report_Exp_%s_%s.xlsx", exp.ExperimentNumber, time.Now().Format("2006-01-02"))

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	if err := f.Write(c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Ошибка при сборке файла отчета")
	}
}
