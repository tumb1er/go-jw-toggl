package store

import (
	"errors"
	"gorm.io/gorm"
	"log"
)

func Must(res *gorm.DB) int64 {
	if res.Error != nil {
		log.Fatalf("%s %T\n", res.Error, res.RowsAffected)
	}
	return res.RowsAffected
}

func MustOrNotFound(res *gorm.DB) bool {
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Fatal(res.Error)
	}
	return res.RowsAffected > 0
}
