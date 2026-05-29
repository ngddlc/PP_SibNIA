package database

import (
	"fmt"
	"log"
	"os"

	"pp_sibnia/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}

	log.Println("Подключение к PostgreSQL успешно!")

	// Автомиграция ВСЕХ наших таблиц
	DB.AutoMigrate(
		&models.Role{}, &models.User{},
		&models.WindTunnel{}, &models.ModelLA{}, &models.Equipment{}, &models.ExperimentType{},
		&models.Experiment{}, &models.Shift{}, &models.Configuration{},
		&models.Protocol{}, &models.ProtocolData{},
	)
	log.Println("Миграции завершены.")
}
