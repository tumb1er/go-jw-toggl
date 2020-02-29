package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/db"
	"time"
)

type DB interface {
	SyncUser() error
	SyncCurrentEntry() error
	OnCurrentEntry(callback db.OnCurrentEntryCallback)
	OnProjectsChanged(callback db.OnProjectsChangedCallback)
	OnTagsChanged(callback db.OnTagsChangedCallback)
	StartListen(ctx context.Context)
	CreateNewEntry(description string, projectId uint64, tagId uint64, start time.Time) error
}
