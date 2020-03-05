package tools

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
)

func TagSlicesEqual(t1, t2 []string) bool {
	if (t1 == nil) != (t2 == nil) {
		return false
	}
	if t1 != nil && t2 != nil {
		if len(t1) != len(t2) {
			return false
		}
		for i, v := range t1 {
			if v != t2[i] {
				return false
			}
		}
		return true
	}
	return true
}

func EntriesEqual(e1, e2 *gttimeentry.TimeEntry) bool {
	if (e1 == nil) != (e2 == nil) {
		return false
	}
	if e1 != nil && e2 != nil {
		return e1.Id == e2.Id &&
			e1.Description == e2.Description &&
			e1.Pid == e2.Pid &&
			TagSlicesEqual(e1.Tags, e2.Tags) &&
			e1.Start.Equal(e2.Start) &&
			e1.Stop.Equal(e2.Stop)
	}
	// e1 == nil && e2 == nil
	return true
}

func ProjectsEqual(p1 *gtproject.Project, p2 *gtproject.Project) bool {
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

func TagsEqual(t1 *gttag.Tag, t2 *gttag.Tag) bool {
	if (t1 == nil) != (t2 == nil) {
		return false
	}
	if t1 != nil && t2 != nil {
		return t1.Id == t2.Id &&
			t1.WId == t2.WId &&
			t1.Name == t2.Name
	}
	// e1 == nil && e2 == nil
	return true
}
