package dto

import "time"

// Пакет dto описывает интерфейсы, используемые в UI и разделяющие
// состояние БД от представления в приложении.

type Tag interface {
	GetName() string
}

type Tags interface {
	DTO() []Tag // store.TagSet -> []Tag
}

type Project interface {
	GetName() string
	GetHexColor() string
	GetPK() uint
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

//TimeEntrySaver описывает методы, достаточные для создания и редактирования записи
type TimeEntrySaver interface {
	GetPK() uint
	Nil() bool
	GetName() string
	GetStart() time.Time
	GetProjectPK() uint
	GetTags() []string
}

//TimeEntry описывает методы, необходимые для корректного отображения записей в истории
type TimeEntry interface {
	TimeEntrySaver
	GetDuration() time.Duration
	TaskKey() string
	GetProject() Project
	GetStop() time.Time
}

type TimeEntries interface {
	DTO() []TimeEntry // store.TimeEntries -> []TimeEntry
}
