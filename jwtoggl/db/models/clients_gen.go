package models

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (obj *Client) BeforeCreate() error {
	return obj.dump(obj.Client)
}
func (obj *Client) BeforeSave() error {
	return obj.dump(obj.Client)
}

func (obj *Client) AfterFind() error {
	return obj.load(&obj.Client)
}

func NewClients(items ...gtclient.Clients) *Clients {
	if len(items) == 0 {
		result := make(Clients, 0)
		return &result
	}
	result := make(Clients, len(items[0]))
	for i, c := range items[0] {
		result[i] = Client{Client: c}
	}
	return &result
}

func (qs Clients) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewClients()
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Client)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.SavedAt = ts
			if prev, ok := extIdMap[obj.Id]; ok {
				obj.PK = prev.PK
				if !tools.ClientsEqual(&obj.Client, &prev.Client) {
					changed = true
				}
				tx.Save(&obj)
			} else {
				changed = true
				tx.Create(&obj)
			}
		}
		res := tx.Where("saved_at < ?", ts).Delete(Client{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}
