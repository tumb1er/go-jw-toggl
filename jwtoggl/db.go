package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/db"
)

type DB interface {
	SyncProjects() error
	SyncCurrentEntry() error
	OnCurrentEntry(callback db.OnCurrentEntryCallback)
	OnProjectsChanged(callback db.OnProjectsChangedCallback)
	StartListen(ctx context.Context)
}
