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
		title: entry.GetName(),
	}
	return ds.tasks[key]
}

type dayGroup map[string]*daySummary

func (dg dayGroup) add(s *gtk.TreeStore, entry dto.TimeEntry) *daySummary {
	ts := entry.GetStart().Local()
	day := ts.Format("2006-01-02")
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
