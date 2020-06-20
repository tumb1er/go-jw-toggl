package entry_form

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"strings"
	"time"
)

type entry struct {
	Description string
	PId         uint64
	Tag         string
	Start       *time.Time
}

func (e entry) ValidateDescription() bool {
	return strings.HasPrefix(e.Description, "#")
}

func (e entry) ValidateTag() bool {
	return e.Tag != ""
}

func (e entry) ValidateProject() bool {
	return e.PId != 0
}

func (e entry) ValidateStart() bool {
	return e.Start != nil
}

func (e entry) Validate() bool {
	return e.ValidateDescription() &&
		e.ValidateProject() &&
		e.ValidateTag() &&
		e.ValidateStart()
}

func (e entry) TimeEntry() *gttimeentry.TimeEntry {
	if !e.Validate() {
		return nil
	}
	return &gttimeentry.TimeEntry{
		Description: e.Description,
		Pid:         e.PId,
		Tags:        []string{e.Tag},
		Start:       *e.Start,
	}
}
