package store

// gtclient.tpl предоставляет методы для работы с моделями объектов Toggl

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

func (m *ClientManager) Sync(items gtclient.Clients, ts time.Time, wid uint64) (bool, error) {
	filtered := make(gtclient.Clients, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewClients(filtered)
	return qs.Sync(m.gorm(), ts)
}

func (m *ClientManager) gorm() *gorm.DB {
	return (*gorm.DB)(m)
}

func NewClients(items ...gtclient.Clients) *Clients {
	if len(items) == 0 {
		result := make(Clients, 0)
		return &result
	}
	result := make(Clients, len(items[0]))
	for i, item := range items[0] {
		result[i] = Client{Client: item}
	}
	return &result
}

func (qs Clients) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewClients()
		Must(tx.Find(existing))
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
				Must(tx.Save(&obj))
			} else {
				changed = true
				Must(tx.Create(&obj))
			}
		}
		changed = MustOrNotFound(tx.Where("saved_at < ?", ts).Delete(Client{})) || changed
		return nil
	})
	return changed, err
}

func listClients(scope *gorm.DB) gtclient.Clients {
	qs := NewClients()
	length := Must(scope.Find(qs))
	result := make(gtclient.Clients, length)
	for i, obj := range *qs {
		result[i] = obj.Client
	}
	return result
}
