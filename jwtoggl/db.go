package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/db"
)

type DB interface {
	SyncUser() error
	SyncCurrentEntry() error
	OnCurrentEntry(callback db.OnCurrentEntryCallback)
	OnProjectsChanged(callback db.OnProjectsChangedCallback)
	OnTagsChanged(callback db.OnTagsChangedCallback)
	StartListen(ctx context.Context)
}
