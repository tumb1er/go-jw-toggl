package store

import (
	"time"
)

type Synced struct {
	PK       uint `gorm:"column:pk;primary_key"`
	Updated  *time.Time
	Modified *time.Time
	Deleted  *time.Time
	Fetched  *time.Time
}

func (s *Synced) BeforeSave() error {
	if s.Updated != nil {
		s.Modified = s.Updated
	} else {
		ts := time.Now().Local()
		s.Modified = &ts
	}
	return nil
}

func (s Synced) After(ts *time.Time) bool {
	if ts == nil {
		return s.Updated != nil
	}
	if s.Updated == nil {
		return false
	}
	return s.Updated.After(*ts)
}

func (s *Synced) MarkDeleted() {
	ts := time.Now().Local()
	s.Deleted = &ts
}
