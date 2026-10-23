package gtk_history

import (
	"strings"
	"testing"
)

func TestHistoryMarkupInheritsThemeForOrdinaryText(t *testing.T) {
	description := RenderDescription(Ctx{
		Description: "Task",
		Tag:         "tag",
		Project:     "Project",
		Color:       "#12AB34",
		Client:      "Client",
	})
	timings := RenderTimings(Ctx{Start: "10:00", Stop: "11:00", Duration: "1h0m0s"})
	date := RenderDate("2026-10-23")

	for name, markup := range map[string]string{
		"description": description,
		"timings":     timings,
		"date":        date,
	} {
		for _, fixedColour := range []string{"#333333", "#555555", "#AAAAAA"} {
			if strings.Contains(markup, fixedColour) {
				t.Fatalf("%s markup contains fixed colour %s: %s", name, fixedColour, markup)
			}
		}
	}
	if !strings.Contains(description, `foreground="#12AB34"`) {
		t.Fatalf("project colour missing from description markup: %s", description)
	}
}
