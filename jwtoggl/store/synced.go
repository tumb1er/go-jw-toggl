package store

import (
	"time"
)

type Synced struct {
	PK uint `gorm:"column:pk;primary_key"`
	//CreatedAt time.Time
	// UpdatedAt time.Time
	SavedAt time.Time
	//FetchedAt time.Time
}

func (s *Synced) BeforeCreate() error {
	s.SavedAt = time.Now().Local()
	return nil
}

func (s *Synced) BeforeSave() error {
	s.SavedAt = time.Now().Local()
	return nil
}

//
//func (s *Synced) dump(obj interface{}) error {
//	if buf, err := json.Marshal(obj); err != nil {
//		return err
//	} else {
//		s.DTO = buf
//		s.ExtId = reflect.ValueOf(obj).FieldByName("Id").Uint()
//	}
//	return nil
//}
//
//func (s *Synced) load(obj interface{}) error {
//	return json.Unmarshal(s.DTO, &obj)
//}
