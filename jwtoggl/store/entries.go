package store

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type TimeEntry struct {
	Synced
	gttimeentry.TimeEntry
	Project   Project `gorm:"foreignkey:ProjectPK"`
	ProjectPK uint
	TagSet    []Tag `gorm:"many2many:entry_tags"`
}

type TimeEntries []TimeEntry

func (qs *TimeEntries) Preload(db *gorm.DB) error {
	projectsMap := make(map[uint64]Project)
	projects := NewProjects()
	if err := db.Session(&gorm.Session{}).Find(projects).Error; err != nil {
		return err
	}
	for _, project := range *projects {
		projectsMap[project.Id] = project
	}
	tagsMap := make(map[string]Tag)
	tags := NewTags()
	if err := db.Session(&gorm.Session{}).Find(tags).Error; err != nil {
		return err
	}
	for _, tag := range *tags {
		tagsMap[tag.Name] = tag
	}
	for i, entry := range *qs {
		entry.Project = projectsMap[entry.Pid]
		entry.ProjectPK = entry.Project.PK
		tags := make([]Tag, 0, len(entry.Tags))
		for _, tag := range entry.Tags {
			tags = append(tags, tagsMap[tag])
		}
		entry.TagSet = tags
		(*qs)[i] = entry
	}
	return nil
}

func (obj *TimeEntry) BeforeSave(db *gorm.DB) error {
	obj.loadProjectFromToggl(db)
	obj.loadTagsFromToggl(db)
	return nil
}

func (obj *TimeEntry) AfterFind(_ *gorm.DB) error {
	obj.computeDuration()
	obj.computeTags()
	return nil
}

func (obj *TimeEntry) loadProjectFromToggl(db *gorm.DB) {
	qs := db
	if obj.Project.PK == 0 {
		log.Println("WARNING: Project not preloaded")
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

func (obj *TimeEntry) loadTagsFromToggl(db *gorm.DB) {
	if len(obj.TagSet) == 0 {
		log.Println("WARNING: Tags not preloaded")
	}
	tagsMap := make(map[string]Tag)
	tags := make([]Tag, 0)
	Must(db.Find(&tags))
	for _, t := range tags {
		tagsMap[t.Name] = t
	}
	obj.TagSet = make([]Tag, 0, len(obj.TimeEntry.Tags))
	for _, t := range obj.TimeEntry.Tags {
		obj.TagSet = append(obj.TagSet, tagsMap[t])
	}
}

func (obj TimeEntry) GetProjectPK() uint {
	return obj.ProjectPK
}

func (obj TimeEntry) GetTags() []string {
	if obj.TagSet == nil {
		panic("Tags not preloaded")
	}
	result := make([]string, len(obj.TagSet))
	for i, t := range obj.TagSet {
		result[i] = t.Name
	}
	return result
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
	key += gtoggl.Tag(obj.GetTags())
	return key
}

func (obj TimeEntry) GetProject() dto.Project {
	if obj.Project.PK == 0 {
		if obj.ProjectPK != 0 {
			panic("project not preloaded")
		} else {
			log.Println("WARNING: project id zero")
		}
	}
	return obj.Project
}

func (obj TimeEntry) GetStop() time.Time {
	return obj.Stop
}

func (obj *TimeEntry) computeDuration() {
	if obj.Stop.IsZero() {
		obj.Duration = -obj.Start.Unix()
	} else {
		obj.Duration = obj.Stop.Unix() - obj.Start.Unix()
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
		if !gtoggl.TagSlicesEqual(obj.GetTags(), other.GetTags()) {
			log.Printf("Tag test: %s %s\n", obj.GetTags(), other.GetTags())
		}
		if !obj.Start.Equal(other.Start) {
			log.Printf("Start test: %s %s\n", obj.Start, other.Start)
		}
		if !obj.Stop.Equal(other.Stop) {
			log.Printf("Stop test: %s %s\n", obj.Stop, other.Stop)
		}
	}
	return obj.Id == other.Id &&
		obj.Name == other.Name &&
		obj.Pid == other.Pid &&
		gtoggl.TagSlicesEqual(obj.GetTags(), other.GetTags()) &&
		obj.Start == other.Start &&
		obj.Stop == other.Stop
}

func (obj *TimeEntry) computeTags() {
	obj.TimeEntry.Tags = obj.GetTags()
}

func (obj *TimeEntry) MarkFetched() {
	ts := time.Now().Local()
	obj.Fetched = &ts
}

func (obj *TimeEntry) MarkModified() {
	ts := time.Now().Local()
	obj.Modified = &ts
}

func (obj *TimeEntry) SetTimeEntry(entry dto.TimeEntrySaver, project Project) {
	obj.Project = project
	obj.ProjectPK = project.PK
	obj.Pid = project.Id

	obj.Name = entry.GetName()
	obj.Start = entry.GetStart()
	obj.Tags = entry.GetTags()
	obj.computeDuration()
}

func (obj *TimeEntry) MarkStopped() {
	ts := time.Now().Local()
	obj.Stop = ts
	obj.Modified = &ts
	obj.computeDuration()
}
