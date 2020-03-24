package gtk_history

import (
	"bytes"
	"log"
	"text/template"
)

var descriptionTemplate, timingTemplate, dateTemplate *template.Template

func init() {
	descriptionTemplate = new(template.Template)
	descriptionTemplate = template.Must(descriptionTemplate.Parse(
		`<span
foreground="#333333">{{.Description}}</span>
{{ if .Tag }}<span 
size="smaller"
font_style="italic">{{.Tag}}</span> {{ end }}{{ if .Project}}<span 
font_weight="bold" 
size="smaller"
foreground="{{.Color}}">• {{.Project}}</span>{{ if .Client }} <span
size="smaller"
foreground="#AAAAAA">• {{.Client}}</span>{{ end }}{{ end }}`))
	timingTemplate = new(template.Template)
	timingTemplate = template.Must(timingTemplate.Parse(
		`<span
	foreground="#AAAAAA">{{.Start}}{{ if .Stop }} - {{.Stop}}{{ else }}...{{ end }}</span>
<span size="smaller">{{.Duration}}</span>`))
	dateTemplate = new(template.Template)
	dateTemplate = template.Must(dateTemplate.Parse(
		`<span
	font_weight="bold"
	foreground="#555555">{{.}}</span>`))
}

type Date struct {
	Date  string
	Total string
}

type Ctx struct {
	Description string
	Color       string
	Project     string
	Client      string
	Tag         string
	Start       string
	Stop        string
	Duration    string
}

func RenderDescription(c Ctx) string {
	buffer := bytes.Buffer{}
	if err := descriptionTemplate.Execute(&buffer, c); err != nil {
		log.Fatal(err)
	}
	return buffer.String()
}
func RenderTimings(c Ctx) string {
	buffer := bytes.Buffer{}
	if err := timingTemplate.Execute(&buffer, c); err != nil {
		log.Fatal(err)
	}
	return buffer.String()
}
func RenderDate(v string) string {
	buffer := bytes.Buffer{}
	if err := dateTemplate.Execute(&buffer, v); err != nil {
		log.Fatal(err)
	}
	return buffer.String()
}
