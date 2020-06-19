package dto

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
	GetClientTogglID() uint64
}

type Projects interface {
	DTO() []Project // store.Projects -> []Project
}
