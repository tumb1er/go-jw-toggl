package dto

// Пакет dto описывает интерфейсы, используемые в UI и разделяющие
// состояние БД от представления в приложении.

type Tag interface {
	GetName() string
}

type Tags interface {
	DTO() []Tag // store.Tags -> []Tag
}
