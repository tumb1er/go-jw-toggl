package store

import (
	"github.com/jinzhu/gorm"
)

type Store gorm.DB

func (s *Store) gorm() *gorm.DB {
	return (*gorm.DB)(s)
}

func (s *Store) Migrate() {
	// Миграция схем
	s.gorm().AutoMigrate(&Tag{}, &Client{}, &Project{}, &TimeEntry{})
}

// TimeEntries возвращает менеджер TimeEntry
func (s *Store) Entries() *TimeEntryManager {
	return (*TimeEntryManager)(s.gorm())
}

func (s *Store) Tags() *TagManager {
	return (*TagManager)(s.gorm())
}

func (s *Store) Projects() *ProjectManager {
	return (*ProjectManager)(s.gorm())
}

func (s *Store) Clients() *ClientManager {
	return (*ClientManager)(s.gorm())
}
