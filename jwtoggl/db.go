package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/db"
	gttimeentry "go-jw-toggl/jwtoggl/gtoggl/gttimentry"
)

type DB interface {
	SyncUser() error
	SyncCurrentEntry() error
	OnCurrentEntry(callback db.OnCurrentEntryCallback)
	OnProjectsChanged(callback db.OnProjectsChangedCallback)
	OnTagsChanged(callback db.OnTagsChangedCallback)
	OnClientsChanged(callback db.OnClientsChangedCallback)
	OnEntriesChanged(callback db.OnEntriesChangedCallback)
	StartListen(ctx context.Context)
	CreateNewEntry(entry *gttimeentry.TimeEntry) error
	StopEntry(entry *gttimeentry.TimeEntry) error
	UpdateCurrentEntry(entry *gttimeentry.TimeEntry) error
	SyncLastEntries() error
	ContinueEntry(id uint64) error
	DeleteEntry(id uint64) error
}
