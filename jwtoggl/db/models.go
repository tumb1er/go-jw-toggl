package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
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

func (tags *Tags) Sync(db *gorm.DB, fetchedAt time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
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
				tx.Save(&t)
			} else {
				tx.Create(&t)
			}
		}
		return tx.Where("fetched_at < ?", fetchedAt).Delete(Tag{}).Error
	})

}
