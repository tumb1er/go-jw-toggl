package entry_form

import (
	"go-jw-toggl/jwtoggl/dto"
	"strings"
	"time"
)

type entry struct {
	Description string
	ProjectPK   uint
	Tag         string
	Start       *time.Time
	PK          uint
}

type project struct {
	PK uint
}

func (p project) GetName() string {
	panic("implement me")
}

func (p project) GetHexColor() string {
	panic("implement me")
}

func (p project) GetPK() uint {
	return p.PK
}

func (p project) GetClient() dto.Client {
	panic("implement me")
}

func (e entry) GetProject() dto.Project {
	return project{e.ProjectPK}
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
	return (*e.Start).Local()
}

func (e entry) TaskKey() string {
	panic("implement me")
}

func (e entry) GetName() string {
	return e.Description
}

func (e entry) GetTag() string {
	return e.Tag
}

func (e entry) GetStop() time.Time {
	return time.Time{}
}

func (e entry) GetPK() uint {
	return e.PK
}

func (e entry) Nil() bool {
	return e.PK == 0
}

func (e entry) ValidateDescription() bool {
	return strings.HasPrefix(e.Description, "#")
}

func (e entry) ValidateTag() bool {
	return e.Tag != ""
}

func (e entry) ValidateProject() bool {
	return e.ProjectPK != 0
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
