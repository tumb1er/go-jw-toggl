package jwtoggl

import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/history"
	"log"
)

func (j *jApp) initHistory(b *builder.Builder) {
	var err error
	if j.entries, err = history.New(b); err != nil {
		log.Fatal(err)
	}
	j.entries.InitCallbacks(j.OnEntryContinue, j.OnEntryDelete)
}

func (j *jApp) OnEntryContinue(pk uint) {
	j.handleError(j.db.ContinueEntry(pk), "Continue entry error")
}

func (j *jApp) OnEntryDelete(pk uint) {
	j.handleError(j.db.DeleteEntry(pk), "Entry delete error")
}
