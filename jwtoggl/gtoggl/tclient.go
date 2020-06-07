package gtoggl

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"go-jw-toggl/jwtoggl/gtoggl/gtworkspace"
)

// Client is an Toggl REST client. Created by calling NewClient.
type TogglClient struct {
	TogglHttpClient *gthttp.TogglHttpClient
	WorkspaceClient *gtworkspace.WorkspaceClient
	ProjectClient   *gtproject.ProjectClient
	TClient         *gtclient.TClient
	UserClient      *gtuser.UserClient
	TimeentryClient *gttimeentry.TimeEntryClient
	TagClient       *gttag.TagClient
}

// Return a new TogglHttpClient . An error is also returned when some configuration option is invalid
//    tc,err := gtoggl.NewClient("token")
func NewClient(key string, options ...gthttp.ClientOptionFunc) (*TogglClient, error) {
	// Set up the client
	c, err := gthttp.NewClient(key, options...)
	if err != nil {
		return nil, err
	}
	th := &TogglClient{TogglHttpClient: c,
		WorkspaceClient: gtworkspace.NewClient(c),
		UserClient:      gtuser.NewClient(c),
		ProjectClient:   gtproject.NewClient(c),
		TagClient:       gttag.NewClient(c),
		TClient:         gtclient.NewClient(c),
		TimeentryClient: gttimeentry.NewClient(c),
	}
	// Run the options on it
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	return th, nil
}
