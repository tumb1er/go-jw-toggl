package jwtoggl

import gttimeentry "github.com/tumb1er/gtoggl-api/gttimentry"

type DB interface {
	SyncCurrentEntry() error
	GetCurrentEntry() *gttimeentry.TimeEntry
}
