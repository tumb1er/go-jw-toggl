package db

import (
	"github.com/tumb1er/gtoggl-api/gtproject"
	"github.com/tumb1er/gtoggl-api/gttimentry"
)

func entriesEqual(e1 *gttimeentry.TimeEntry, e2 *gttimeentry.TimeEntry) bool {
	if (e1 == nil) != (e2 == nil) {
		return false
	}
	if e1 != nil && e2 != nil {
		return e1.Id == e2.Id &&
			e1.Description == e2.Description &&
			e1.Start.Equal(e2.Start) &&
			e1.Stop.Equal(e2.Stop)
	}
	// e1 == nil && e2 == nil
	return true
}

func projectsEqual(p1 *gtproject.Project, p2 *gtproject.Project) bool {
	if (p1 == nil) != (p2 == nil) {
		return false
	}
	if p1 != nil && p2 != nil {
		return p1.Id == p2.Id &&
			p1.CId == p2.CId &&
			p1.Name == p2.Name &&
			p1.WId == p2.WId
	}
	// e1 == nil && e2 == nil
	return true
}
