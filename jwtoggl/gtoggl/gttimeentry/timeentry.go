package gttimeentry

import (
	"encoding/json"
	"fmt"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"net/url"
	"time"
)

type TimeEntry struct {
	Id          uint64    `json:"id,omitempty"`
	Name        string    `json:"description"`
	Start       time.Time `json:"start"`
	Stop        time.Time `json:"stop"`
	Duration    int64     `json:"duration" gorm:"-"`
	Billable    bool      `json:"billable"`
	Tags        []string  `json:"tags" gorm:"-"`
	Pid         uint64    `json:"pid"`
	WId         uint64    `json:"wid"`
	Tid         uint64    `json:"tid"`
	CreatedWith string    `json:"created_with,omitempty" gorm:"-"`
	At          time.Time `json:"at,omitempty" gorm:"-"`
}

type TimeEntries []TimeEntry

const Endpoint = "/me/time_entries"
const EndpointCurrent = Endpoint + "/current"
const EndpointStart = Endpoint + "/start"
const EndpointCreate = "/workspaces/%d/time_entries"
const EndpointUpdate = "/workspaces/%d/time_entries/%d"

// Return a UserClient. An error is also returned when some configuration option is invalid
//
//	thc,err := gtoggl.NewClient("token")
//	uc,err := guser.NewClient(thc)
func NewClient(thc *gthttp.TogglHttpClient) *TimeEntryClient {
	tc := &TimeEntryClient{
		thc: thc,
	}
	tc.endpoint = thc.Url + Endpoint
	tc.currentEndpoint = thc.Url + EndpointCurrent
	tc.startEndpoint = thc.Url + EndpointStart
	tc.createEndpoint = thc.Url + EndpointCreate
	tc.updateEndpoint = thc.Url + EndpointUpdate
	return tc
}

type TimeEntryClient struct {
	thc             *gthttp.TogglHttpClient
	endpoint        string
	startEndpoint   string
	currentEndpoint string
	createEndpoint  string
	updateEndpoint  string
}

func (c *TimeEntryClient) Get(tid uint64) (*TimeEntry, error) {
	return timeEntryResponse(c.thc.GetRequest(fmt.Sprintf("%s/%d", c.endpoint, tid)))
}

func (c *TimeEntryClient) Current() (*TimeEntry, error) {
	return timeEntryResponse(c.thc.GetRequest(c.currentEndpoint))
}

func (c *TimeEntryClient) Delete(id uint64) error {
	_, err := c.thc.DeleteRequest(fmt.Sprintf("%s/%d", c.endpoint, id), nil)
	return err
}

func (c *TimeEntryClient) List() (TimeEntries, error) {
	body, err := c.thc.GetRequest(c.endpoint)
	var te TimeEntries
	if err != nil {
		return te, err
	}
	if body == nil {
		return nil, nil
	}
	err = json.Unmarshal(*body, &te)
	return te, err
}

func (c *TimeEntryClient) Create(t *TimeEntry) (*TimeEntry, error) {
	if len(t.CreatedWith) < 0 {
		t.CreatedWith = "gtoggl"
	}

	type TimeEntryCreation struct {
		Id          uint64     `json:"id,omitempty"`
		Name        string     `json:"description"`
		Start       time.Time  `json:"start"`
		Stop        *time.Time `json:"stop,omitempty"`
		Duration    int64      `json:"duration" gorm:"-"`
		Billable    bool       `json:"billable"`
		Tags        []string   `json:"tags" gorm:"-"`
		Pid         uint64     `json:"pid"`
		WId         uint64     `json:"wid"`
		Tid         *uint64    `json:"tid"`
		CreatedWith string     `json:"created_with,omitempty" gorm:"-"`
		At          time.Time  `json:"at,omitempty" gorm:"-"`
	}
	dto := TimeEntryCreation{
		Id:       t.Id,
		Name:     t.Name,
		Start:    t.Start,
		Duration: t.Duration,
		Billable: t.Billable,
		Tags:     t.Tags,
		Pid:      t.Pid,
		WId:      t.WId,
		// Tid - empty
		CreatedWith: t.CreatedWith,
		At:          t.At,
	}
	if !t.Stop.IsZero() {
		dto.Stop = &t.Stop
	}
	return timeEntryResponse(c.thc.PostRequest(fmt.Sprintf(c.createEndpoint, dto.WId), dto))
}

func (c *TimeEntryClient) Start(t *TimeEntry) (*TimeEntry, error) {
	if len(t.CreatedWith) < 0 {
		t.CreatedWith = "gtoggl"
	}
	up := map[string]interface{}{"time_entry": t}
	return timeEntryResponse(c.thc.PostRequest(c.startEndpoint, up))
}

func (c *TimeEntryClient) Stop(t *TimeEntry) (*TimeEntry, error) {
	return timeEntryResponse(c.thc.PutRequest(fmt.Sprintf("%s/%d/stop", c.endpoint, t.Id), nil))
}

func (c *TimeEntryClient) Update(t *TimeEntry) (*TimeEntry, error) {
	up := map[string]interface{}{"time_entry": t}
	return timeEntryResponse(c.thc.PutRequest(fmt.Sprintf(c.updateEndpoint, t.WId, t.Id), up))
}

func (c *TimeEntryClient) GetRange(start time.Time, end time.Time) (TimeEntries, error) {
	v := url.Values{}
	v.Set("start_date", start.Format(time.RFC3339))
	v.Set("end_date", end.Format(time.RFC3339))
	body, err := c.thc.GetRequest(fmt.Sprintf("%s?%s", c.endpoint, v.Encode()))
	var te TimeEntries
	if err != nil {
		return te, err
	}
	if body == nil {
		return nil, nil
	}
	err = json.Unmarshal(*body, &te)
	return te, err
}

func timeEntryResponse(response *json.RawMessage, error error) (*TimeEntry, error) {
	if error != nil {
		return nil, error
	}
	if response == nil || len(*response) == 4 {
		return nil, nil
	}
	var t TimeEntry
	err := json.Unmarshal(*response, &t)
	if err != nil {
		return nil, err
	}
	return &t, err
}
