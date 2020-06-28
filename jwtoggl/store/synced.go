package store

import (
	"time"
)

type Synced struct {
	PK        uint `gorm:"column:pk;primary_key"`
	SavedAt   time.Time
	FlaggedAt time.Time
}

func (s *Synced) BeforeSave() error {
	s.SavedAt = time.Now().Local()
	return nil
}
