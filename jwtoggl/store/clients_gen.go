package store

// gtclient.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

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
	return (*gorm.DB)(m).Set("gorm:auto_preload", true)
}

func NewClients(items ...gtclient.Clients) *Clients {
	if len(items) == 0 {
		result := make(Clients, 0)
		return &result
	}
	result := make(Clients, len(items[0]))
	for i, item := range items[0] {
		result[i] = Client{Client: item}
		result[i].ComputeUpdated()
	}
	return &result
}

func (qs Clients) DTO() []dto.Client {
	result := make([]dto.Client, len(qs))
	for i, c := range qs {
		result[i] = c
	}
	return result
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
			if prev, ok := extIdMap[obj.Id]; ok {
				if prev.After(obj.Updated) {
					// log.Printf("Skip saving Client %d %s\n\n", obj.Id, obj.Name)
					obj.Client = prev.Client
				}
				obj.PK = prev.PK
				if !tools.ClientsEqual(&obj.Client, &prev.Client) {
					changed = true
				}
				Must(tx.Debug().Save(&obj))
			} else {
				changed = true
				Must(tx.Debug().Create(&obj))
			}
		}
		changed = MustOrNotFound(tx.Debug().Where("modified < updated").Delete(Client{})) || changed
		return nil
	})
	return changed, err
}

func listClients(scope *gorm.DB) Clients {
	qs := NewClients()
	Must(scope.Debug().Find(qs))
	return *qs
}

func (obj Client) GetName() string {
	return obj.Name
}

func (obj Client) GetPK() uint {
	return obj.PK
}

func (obj *Client) ComputeUpdated() {
	obj.Updated = &obj.Client.At
}
