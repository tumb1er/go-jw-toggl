package gtk_history

import (
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/dto"
	"time"
)

type taskSummary struct {
	it    *gtk.TreeIter
	title string
	count int
}

type daySummary struct {
	it       *gtk.TreeIter
	duration time.Duration
	tasks    map[string]*taskSummary
}

func (ds *daySummary) add(s *gtk.TreeStore, entry dto.TimeEntry) *taskSummary {
	key := entry.TaskKey()
	summary, ok := ds.tasks[key]
	if ok {
		return summary
	}
	ds.tasks[key] = &taskSummary{
		it:    s.Append(ds.it),
		title: entry.GetDescription(),
	}
	return ds.tasks[key]
}

type dayGroup map[time.Time]*daySummary

func (dg dayGroup) add(s *gtk.TreeStore, entry dto.TimeEntry) *daySummary {
	day := entry.GetStart().Local().Truncate(24 * time.Hour)
	summary, ok := dg[day]
	if ok {
		return summary
	}
	dg[day] = &daySummary{
		it:    s.Append(nil),
		tasks: make(map[string]*taskSummary),
	}
	return dg[day]
}
