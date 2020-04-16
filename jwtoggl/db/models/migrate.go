package models

import "github.com/jinzhu/gorm"

func Migrate(db *gorm.DB) {
	// Миграция схем
	db.AutoMigrate(&Tag{}, &Client{}, &Project{}, &TimeEntry{})
}
