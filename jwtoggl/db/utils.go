package db

import "github.com/tumb1er/gtoggl-api/gttimentry"

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
