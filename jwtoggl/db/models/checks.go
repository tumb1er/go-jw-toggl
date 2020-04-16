package models

import (
	"github.com/jinzhu/gorm"
	"log"
)

func Must(res *gorm.DB) int64 {
	if res.Error != nil {
		log.Fatalf("%s %T\n", res.Error, res.Value)
	}
	return res.RowsAffected
}

func MustOrNotFound(res *gorm.DB) bool {
	if res.Error != nil && !res.RecordNotFound() {
		log.Fatal(res.Error)
	}
	return res.RowsAffected > 0
}
