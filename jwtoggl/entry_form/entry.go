package entry_form

import (
	"go-jw-toggl/jwtoggl/dto"
	"strings"
	"time"
)

type entry struct {
	Description string
	PId         uint64
	Tag         string
	Start       *time.Time
	Id          uint64
}

func (e entry) GetTags() []string {
	if e.Tag == "" {
		return []string{}
	}
	return []string{e.Tag}
}

func (e entry) GetDuration() time.Duration {
	panic("implement me")
}

func (e entry) GetStart() time.Time {
	if e.Start == nil {
		return time.Time{}
	}
	return *e.Start
}

func (e entry) TaskKey() string {
	panic("implement me")
}

func (e entry) GetName() string {
	return e.Description
}

func (e entry) GetProjectTogglID() uint64 {
	return e.PId
}

func (e entry) GetTag() string {
	return e.Tag
}

func (e entry) GetStop() time.Time {
	return time.Time{}
}

func (e entry) GetTogglID() uint64 {
	return e.Id
}

func (e entry) Nil() bool {
	return true
}

func (e entry) Equals(entry dto.TimeEntry) bool {
	panic("implement me")
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
