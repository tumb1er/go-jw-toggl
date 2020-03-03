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
	StartListen(ctx context.Context)
	CreateNewEntry(entry *gttimeentry.TimeEntry) error
}
