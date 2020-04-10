package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

type Synced struct {
	PK        uint `gorm:"column:pk;primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	FetchedAt time.Time
}

type Tag struct {
	Synced
	gttag.Tag
}

type Tags []Tag

func NewTags(tags gttag.Tags) *Tags {
	queryset := make(Tags, len(tags))
	for i, t := range tags {
		queryset[i] = Tag{Tag: t}
	}
	return &queryset
}

func (tags *Tags) Sync(db *gorm.DB, fetchedAt time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := new(Tags)
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Tag)
		for _, t := range *existing {
			extIdMap[t.Id] = t
		}

		for _, t := range *tags {
			t.FetchedAt = fetchedAt
			if e, ok := extIdMap[t.Id]; ok {
				t.PK = e.PK
				if !tools.TagsEqual(&t.Tag, &e.Tag) {
					changed = true
				}
				tx.Save(&t)
			} else {
				changed = true
				tx.Create(&t)
			}
		}
		res := tx.Where("fetched_at < ?", fetchedAt).Delete(Tag{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}

type Client struct {
	Synced
	gtclient.Client
}

type Clients []Client

func NewClients(clients gtclient.Clients) *Clients {
	queryset := make(Clients, len(clients))
	for i, c := range clients {
		queryset[i] = Client{Client: c}
	}
	return &queryset
}

func (clients *Clients) Sync(db *gorm.DB, fetchedAt time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := new(Clients)
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Client)
		for _, c := range *existing {
			extIdMap[c.Id] = c
		}

		for _, c := range *clients {
			c.FetchedAt = fetchedAt
			if e, ok := extIdMap[c.Id]; ok {
				c.PK = e.PK
				if !tools.ClientsEqual(&c.Client, &e.Client) {
					changed = true
				}
				tx.Save(&c)
			} else {
				changed = true
				tx.Create(&c)
			}
		}
		res := tx.Where("fetched_at < ?", fetchedAt).Delete(Client{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}
