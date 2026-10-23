package state

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type recordingTransport struct {
	requests []string
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "api.track.toggl.com" {
		return nil, fmt.Errorf("unexpected host: %s", req.URL.Host)
	}
	t.requests = append(t.requests, req.Method+" "+req.URL.Path)

	var body string
	switch req.Method {
	case http.MethodPut:
		body = `{"id":10,"description":"previous","start":"2026-10-23T10:00:00Z","stop":"2026-10-23T11:00:00Z","duration":3600,"pid":100,"wid":1,"at":"2026-10-23T11:00:00Z"}`
	case http.MethodPost:
		body = `{"id":11,"description":"next","start":"2026-10-23T11:00:00Z","duration":-1,"pid":100,"wid":1,"at":"2026-10-23T11:00:00Z"}`
	default:
		return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func TestSyncChangesUpdatesPreviousEntryBeforeCreatingNext(t *testing.T) {
	temporaryDirectory, err := ioutil.TempDir("", "go-jw-toggl-state-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(temporaryDirectory)

	gormDB, err := gorm.Open(sqlite.Open(filepath.Join(temporaryDirectory, "toggl.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	localStore := &StoreAdapter{Store: (*store.Store)(gormDB)}
	db, err := NewDB("offline-test-token", localStore)
	if err != nil {
		t.Fatal(err)
	}
	db.defaultWId = 1
	transport := &recordingTransport{}
	db.httpClient.Transport = transport
	db.OnCurrentEntry(func(_ dto.TimeEntry, _ bool) {})
	db.onTimeEntriesChanged = func(_ dto.TimeEntries) {}

	project := store.Project{Project: gtproject.Project{Id: 100, WId: 1, Name: "Project"}}
	if err := gormDB.Create(&project).Error; err != nil {
		t.Fatal(err)
	}

	start := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	stop := start.Add(time.Hour)
	previous := store.TimeEntry{
		TimeEntry: gttimeentry.TimeEntry{Id: 10, Name: "previous", Start: start, Pid: project.Id, WId: 1, At: start},
		Project:   project,
		ProjectPK: project.PK,
		TagSet:    []store.Tag{},
	}
	previous.ComputeUpdated()
	previous.Modified = previous.Updated
	if err := gormDB.Create(&previous).Error; err != nil {
		t.Fatal(err)
	}

	next := store.TimeEntry{
		TimeEntry: gttimeentry.TimeEntry{Name: "next", Start: stop, Duration: -1, Pid: project.Id, WId: 1},
		Project:   project,
		ProjectPK: project.PK,
		TagSet:    []store.Tag{},
	}
	if err := db.CreateEntry(next); err != nil {
		t.Fatal(err)
	}

	if err := db.SyncChanges(); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"PUT /api/v9/workspaces/1/time_entries/10",
		"POST /api/v9/workspaces/1/time_entries",
	}
	if !reflect.DeepEqual(transport.requests, want) {
		t.Fatalf("requests = %v, want %v", transport.requests, want)
	}
}
