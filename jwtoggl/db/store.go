package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"sort"
	"sync"
)

type EntryStore struct {
	index map[uint64]int
	list  []gttimeentry.TimeEntry
	mu    sync.RWMutex
}

func (s *EntryStore) Update(entry gttimeentry.TimeEntry) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := s.update(entry)
	if updated {
		s.sort()
	}
	return updated
}

func (s *EntryStore) update(entry gttimeentry.TimeEntry) bool {
	if entry.Id == 0 {
		log.Fatal("empty entry id")
	}
	idx, ok := s.index[entry.Id]
	if !ok {
		// Отсутствует в индексе
		s.list = append(s.list, entry)
		s.sort()
		return true
	}
	if tools.EntriesEqual(&s.list[idx], &entry) {
		// Нет изменений
		return false
	}
	s.list[idx] = entry
	s.sort()
	return true
}

func (s *EntryStore) sort() {
	sort.Slice(s.list, func(i, j int) bool {
		return s.list[i].Start.After(s.list[j].Start)
	})
	s.reindex()
}

func (s *EntryStore) reindex() {
	s.index = make(map[uint64]int, len(s.list))
	for i, v := range s.list {
		s.index[v.Id] = i
	}
}
