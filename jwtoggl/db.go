package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/db"
)

type DB interface {
	SyncCurrentEntry() error
	OnCurrentEntry(callback db.OnCurrentEntryCallback)
	StartListen(ctx context.Context)
}
