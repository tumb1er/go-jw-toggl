package dto

import "time"

// Пакет dto описывает интерфейсы, используемые в UI и разделяющие
// состояние БД от представления в приложении.

type Tag interface {
	GetName() string
}

type Tags interface {
	DTO() []Tag // store.Tags -> []Tag
}

type Project interface {
	GetName() string
	GetHexColor() string
	GetTogglID() uint64
	GetClient() Client
}

type Projects interface {
	DTO() []Project // store.Projects -> []Project
}

type Client interface {
	GetName() string
}

type Clients interface {
	DTO() []Client // store.Clients -> []Client
}

type TimeEntry interface {
	GetDuration() time.Duration
	GetStart() time.Time
	TaskKey() string
	GetName() string
	GetProject() Project
	GetTag() string
	GetStop() time.Time
	GetTogglID() uint64
	Nil() bool
	Equals(entry TimeEntry) bool
	GetTags() []string
}

type TimeEntries interface {
	DTO() []TimeEntry // store.TimeEntries -> []TimeEntry
}
