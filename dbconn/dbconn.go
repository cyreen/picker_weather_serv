package dbconn

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"picker_weather_serv/models"
)

func ConnectToPostgreSQL(dbHost string, dbUser string, dbPassword string, dbName string, dbPort string) (*gorm.DB, error) {
	dbc := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", dbUser, dbPassword, dbName, dbHost, dbPort)
	db, err := gorm.Open(postgres.Open(dbc), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func FetchStores(db *gorm.DB) ([]models.Store, error) {
	var stores []models.Store
	result := db.Model(&models.Store{}).Select("id_store, number_store_cyreen, longitude, latitude").Find(&stores)
	if result.Error != nil {
		return nil, result.Error
	}

	return stores, nil
}
