package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	gttimeentry "go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
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

type Project struct {
	Synced
	gtproject.Project
	// Client Client `gorm:"foreignkey:CId;association_foreignkey:Id"`
}

type Projects []Project

func NewProjects(projects gtproject.Projects) *Projects {
	queryset := make(Projects, len(projects))
	for i, p := range projects {
		queryset[i] = Project{Project: p}
	}
	return &queryset
}

func (projects *Projects) Sync(db *gorm.DB, fetchedAt time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := new(Projects)
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Project)
		for _, p := range *existing {
			extIdMap[p.Id] = p
		}

		for _, p := range *projects {
			p.FetchedAt = fetchedAt
			if e, ok := extIdMap[p.Id]; ok {
				p.PK = e.PK
				if !tools.ProjectsEqual(&p.Project, &e.Project) {
					changed = true
				}
				tx.Save(&p)
			} else {
				changed = true
				tx.Create(&p)
			}
		}
		res := tx.Where("fetched_at < ?", fetchedAt).Delete(Project{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}

type Entry struct {
	Synced
	gttimeentry.TimeEntry
	Tags []string `gorm:"-"`
}

func (e Entry) Delete(db *gorm.DB) {
	if res := db.Delete(&e); res.Error != nil && !res.RecordNotFound() {
		log.Fatal(res.Error)
	}
}

type Entries []Entry

func NewEntries(entries gttimeentry.TimeEntries) *Entries {
	queryset := make(Entries, len(entries))
	for i, e := range entries {
		queryset[i] = Entry{TimeEntry: e}
	}
	return &queryset
}

func (entries *Entries) Sync(db *gorm.DB, fetchedAt time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := new(Entries)
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Entry)
		for _, e := range *existing {
			extIdMap[e.Id] = e
		}

		for _, v := range *entries {
			v.FetchedAt = fetchedAt
			if e, ok := extIdMap[v.Id]; ok {
				v.PK = e.PK
				if !tools.EntriesEqual(&v.TimeEntry, &e.TimeEntry) {
					changed = true
				}
				tx.Save(&v)
			} else {
				changed = true
				tx.Create(&v)
			}
		}
		res := tx.Where("fetched_at < ?", fetchedAt).Delete(Entry{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}
