package models

import (
	"encoding/json"
	"reflect"
	"time"
)

type Synced struct {
	PK uint `gorm:"column:pk;primary_key"`
	//CreatedAt time.Time
	// UpdatedAt time.Time
	SavedAt time.Time
	//FetchedAt time.Time
	DTO   []byte `gorm:"type:text"` // json with DTO
	ExtId uint64 // external id from DTO
}

func (s *Synced) dump(obj interface{}) error {
	if buf, err := json.Marshal(obj); err != nil {
		return err
	} else {
		s.DTO = buf
		s.ExtId = reflect.ValueOf(obj).FieldByName("Id").Uint()
	}
	return nil
}

func (s *Synced) load(obj interface{}) error {
	return json.Unmarshal(s.DTO, &obj)
}
