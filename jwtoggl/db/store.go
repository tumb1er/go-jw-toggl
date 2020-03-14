package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"sort"
)

type EntryStore struct {
	index map[uint64]int
	list  []*gttimeentry.TimeEntry
}

func (s *EntryStore) Update(entry *gttimeentry.TimeEntry) bool {
	updated := s.update(entry)
	if updated {
		s.sort()
	}
	return updated
}

func (s *EntryStore) BulkUpdate(entries gttimeentry.TimeEntries) bool {
	var result bool
	for _, entry := range entries {
		entry := entry
		if s.update(&entry) {
			result = true
		}
	}
	if result {
		s.sort()
	}
	return result
}

func (s *EntryStore) update(entry *gttimeentry.TimeEntry) bool {
	if entry == nil {
		return false
	}
	idx, ok := s.index[entry.Id]
	if !ok {
		// Отсутствует в индексе
		s.list = append(s.list, entry)
		s.sort()
		return true
	}
	if tools.EntriesEqual(s.list[idx], entry) {
		// Нет изменений
		return false
	}
	s.list[idx] = entry
	return false
}

func (s *EntryStore) Current() *gttimeentry.TimeEntry {
	for _, entry := range s.list {
		if entry.Stop.IsZero() {
			return entry
		}
	}
	return nil
}

func (s *EntryStore) sort() {
	sort.Slice(s.list, func(i, j int) bool {
		return s.list[i].Start.After(s.list[j].Start)
	})
	s.index = make(map[uint64]int, len(s.list))
	for i, v := range s.list {
		s.index[v.Id] = i
	}
}

func (s *EntryStore) Delete(entry *gttimeentry.TimeEntry) bool {
	idx, ok := s.index[entry.Id]
	if !ok {
		// Не найдено
		return false
	}
	// Сдвигаем влево
	copy(s.list[idx:], s.list[idx+1:])
	// Стираем ссылку
	s.list[len(s.list)-1] = nil
	// Меняем длину
	s.list = s.list[:len(s.list)-1]
	return true
}

func (s *EntryStore) List() []*gttimeentry.TimeEntry {
	result := make([]*gttimeentry.TimeEntry, 0, len(s.list))
	current := s.Current()
	for _, entry := range s.list {
		if current != nil && entry.Id == current.Id {
			continue
		}
		result = append(result, entry)
	}
	return result
}
