package store

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"log"
	"strings"
	"time"
)

type TimeEntry struct {
	Synced
	gttimeentry.TimeEntry
	Tag       string
	Project   Project `gorm:"foreignkey:ProjectPK"`
	ProjectPK uint
}

type TimeEntries []TimeEntry

type TimeEntryManager gorm.DB

func (m *TimeEntryManager) FromTimeEntry(entry gttimeentry.TimeEntry) *TimeEntry {
	timeEntry := m.Get(entry.Id)
	if timeEntry == nil {
		timeEntry = &TimeEntry{}
	}
	timeEntry.TimeEntry = entry
	if timeEntry.Project.Id != entry.Pid {
		timeEntry.Project.PK = 0
		Must(m.gorm().Debug().Where("id = ?", entry.Pid).Find(&timeEntry.Project))
		timeEntry.ProjectPK = timeEntry.Project.PK
	}
	return timeEntry
}

func (obj *TimeEntry) BeforeSave(db *gorm.DB) error {
	if err := obj.Synced.BeforeSave(); err != nil {
		return err
	}
	obj.loadProjectFromToggl(db)
	obj.loadTagFromTags()
	return nil
}

func (obj *TimeEntry) AfterFind() error {
	obj.computeTags()
	obj.ComputeDuration()
	return nil
}

func (obj *TimeEntry) loadProjectFromToggl(db *gorm.DB) {
	qs := db.Set("gorm:auto_preload", true)
	if obj.Project.PK == 0 {
		// Загружаем проект из БД по внутреннему ключу
		if obj.ProjectPK != 0 {
			if !MustOrNotFound(qs.Where("pk = ?", obj.ProjectPK).First(&obj.Project)) {
				obj.ProjectPK = 0
			}
		}
	}
	if obj.Project.PK == 0 {
		// По внутреннему ключу проект не найден, ищем через внешний ключ
		if MustOrNotFound(qs.Where("id = ?", obj.Pid).First(&obj.Project)) {
			// Нашли по внешнему ключу, проставляем внутренний
			obj.ProjectPK = obj.Project.PK
		} else if obj.Pid != 0 {
			panic("Project missing")
		}
	}
}

func (obj TimeEntry) GetTags() []string {
	if obj.Tag == "" {
		return []string{}
	}
	return []string{obj.Tag}
}

func (obj TimeEntry) Nil() bool {
	return obj.PK == 0
}

func (obj TimeEntry) IsCurrent() bool {
	return obj.Stop.IsZero()
}

func (obj TimeEntry) GetDuration() time.Duration {
	return obj.Stop.Sub(obj.Start)
}

func (obj TimeEntry) GetStart() time.Time {
	return obj.Start.Local()
}

func (obj TimeEntry) TaskKey() string {
	key := obj.Name
	if strings.HasPrefix(obj.Name, "#") {
		key = strings.SplitN(obj.Name, " ", 2)[0]
	}
	if len(obj.Tags) > 0 {
		key += obj.Tags[0]
	}
	return key
}

func (obj TimeEntry) GetProject() dto.Project {
	if obj.Project.PK == 0 {
		panic("project not preloaded")
	}
	return obj.Project
}

func (obj TimeEntry) GetTag() string {
	if len(obj.Tags) > 0 {
		return obj.Tags[0]
	}
	return ""
}

func (obj TimeEntry) GetStop() time.Time {
	return obj.Stop
}

func (obj *TimeEntry) loadTagFromTags() {
	obj.Tag = ""
	if len(obj.Tags) > 0 {
		obj.Tag = obj.Tags[0]
	}
}

func (obj *TimeEntry) ComputeDuration() {
	if obj.Stop.IsZero() {
		obj.Duration = -obj.Start.Unix()
	} else {
		obj.Duration = obj.Stop.Unix() - obj.Start.Unix()
	}
}

func (obj *TimeEntry) computeTags() {
	if obj.Tag != "" {
		obj.Tags = []string{obj.Tag}
	} else {
		obj.Tags = []string{}
	}
}

func (obj TimeEntry) Equals(other TimeEntry) bool {
	if obj.Nil() || other.Nil() {
		// Не допускаем получения проекта у несуществующей записи.
		return obj.Nil() == other.Nil()
	}

	const equalsDebug = false
	if equalsDebug {
		if obj.Nil() != other.Nil() {
			log.Printf("Nil test: %+v %+v\n", obj.Nil(), other.Nil())
		}
		if obj.Id != other.Id {
			log.Printf("ID test: %d %d\n", obj.Id, other.Id)
		}
		if obj.Name != other.Name {
			log.Printf("Name test: %s %s\n", obj.Name, other.Name)
		}
		if obj.Pid != other.Pid {
			log.Printf("Project test: %d %d\n", obj.Pid, other.Pid)
		}
		if obj.GetTag() != other.GetTag() {
			log.Printf("Tag test: %s %s\n", obj.GetTag(), other.GetTag())
		}
		if obj.Start != other.Start {
			log.Printf("Start test: %s %s\n", obj.Start, other.Start)
		}
		if obj.Stop != other.Stop {
			log.Printf("Stop test: %s %s\n", obj.Stop, other.Stop)
		}
	}
	return obj.Id == other.Id &&
		obj.Name == other.Name &&
		obj.Pid == other.Pid &&
		obj.GetTag() == other.GetTag() &&
		obj.Start == other.Start &&
		obj.Stop == other.Stop
}

func (m *TimeEntryManager) List() TimeEntries {
	return listTimeEntries(
		m.gorm().
			Where("stop <> ?", time.Time{}).
			Order("start DESC"))
}

func (m *TimeEntryManager) Delete(extId uint64) bool {
	return MustOrNotFound(m.gorm().Where("id = ?", extId).Delete(&TimeEntry{}))
}

func (m *TimeEntryManager) Current() TimeEntry {
	var entry TimeEntry
	MustOrNotFound(m.gorm().
		Where("stop = ?", time.Time{}).
		First(&entry))
	return entry
}

func (m *TimeEntryManager) Create(timeEntry *TimeEntry) {
	Must(m.gorm().Create(timeEntry))
}

func (m *TimeEntryManager) Update(timeEntry *TimeEntry) {
	Must(m.gorm().Model(timeEntry).
		Where("pk = ?", timeEntry.PK).
		Updates(timeEntry))
}

func (m *TimeEntryManager) Get(extId uint64) *TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().
		Where("id = ?", extId).First(&entry)) {
		return &entry
	}
	return nil
}

func (m *TimeEntryManager) GetByPK(pk uint) *TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().
		Where("pk = ?", pk).First(&entry)) {
		return &entry
	}
	return nil
}
