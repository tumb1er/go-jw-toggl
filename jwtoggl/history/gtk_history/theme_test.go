package gtk_history

import (
	"testing"

	"github.com/gotk3/gotk3/gdk"
)

func TestDayBackgroundDarkensLightThemeListSurface(t *testing.T) {
	got := dayBackgroundFromTheme(func(name string) (string, bool) {
		switch name {
		case "theme_base_color":
			return "#ffffff", true
		case "theme_fg_color":
			return "#3d3d3d", true
		}
		return "", false
	})
	if want := "#efefef"; got != want {
		t.Fatalf("day background = %q, want %q", got, want)
	}
}

func TestDayBackgroundLightensDarkThemeListSurface(t *testing.T) {
	got := dayBackgroundFromTheme(func(name string) (string, bool) {
		switch name {
		case "theme_base_color":
			return "#242424", true
		case "theme_fg_color":
			return "#ffffff", true
		}
		return "", false
	})
	if want := "#363636"; got != want {
		t.Fatalf("day background = %q, want %q", got, want)
	}
}

func TestDayBackgroundDoesNotInventPaletteWhenThemeColorIsUnavailable(t *testing.T) {
	got := dayBackgroundFromTheme(func(string) (string, bool) { return "", false })
	if got != "" {
		t.Fatalf("day background = %q, want empty", got)
	}
}

func TestCellRendererColorUsesHexFormat(t *testing.T) {
	color := gdk.NewRGBA()
	color.SetRed(1)
	color.SetGreen(0.5)
	color.SetBlue(0)
	color.SetAlpha(1)

	if got, want := cellRendererColor(color), "#ff8000"; got != want {
		t.Fatalf("cell renderer colour = %q, want %q", got, want)
	}
}
