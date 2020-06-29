package entry_form

import (
	"strings"
	"time"
)

type entry struct {
	Name      string
	ProjectPK uint
	Tag       string
	Start     *time.Time
	PK        uint
}

func (e entry) GetProjectPK() uint {
	return e.ProjectPK
}

func (e entry) GetTags() []string {
	if e.Tag == "" {
		return []string{}
	}
	return []string{e.Tag}
}

func (e entry) GetStart() time.Time {
	if e.Start == nil {
		return time.Time{}
	}
	return (*e.Start).Local()
}

func (e entry) GetName() string {
	return e.Name
}

func (e entry) GetPK() uint {
	return e.PK
}

func (e entry) Nil() bool {
	return e.PK == 0
}

func (e entry) ValidateDescription() bool {
	return strings.HasPrefix(e.Name, "#")
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
