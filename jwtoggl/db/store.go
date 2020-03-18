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

func (s *EntryStore) BulkUpdate(entries gttimeentry.TimeEntries) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result bool
	for _, entry := range entries {
		entry := entry
		if s.update(entry) {
			result = true
		}
	}
	if result {
		s.sort()
	}
	return result
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

func (s *EntryStore) Get(id uint64) (gttimeentry.TimeEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, ok := s.index[id]
	if !ok {
		return gttimeentry.TimeEntry{}, false
	}
	return s.list[idx], true
}

func (s *EntryStore) current() (gttimeentry.TimeEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, entry := range s.list {
		if tools.IsCurrent(entry) {
			return entry, true
		}
	}
	return gttimeentry.TimeEntry{}, false
}

func (s *EntryStore) Current() *gttimeentry.TimeEntry {
	entry, ok := s.current()
	if ok {
		return &entry
	} else {
		return nil
	}
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

func (s *EntryStore) Delete(entry gttimeentry.TimeEntry) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, ok := s.index[entry.Id]
	if !ok {
		// Не найдено
		return false
	}
	// Сдвигаем влево
	copy(s.list[idx:], s.list[idx+1:])
	// Стираем данные
	s.list[len(s.list)-1] = gttimeentry.TimeEntry{}
	// Меняем длину
	s.list = s.list[:len(s.list)-1]
	s.reindex()
	return true
}

func (s *EntryStore) List() gttimeentry.TimeEntries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(gttimeentry.TimeEntries, 0, len(s.list))
	current, ok := s.current()
	for _, entry := range s.list {
		if ok && entry.Id == current.Id {
			continue
		}
		result = append(result, entry)
	}
	return result
}
