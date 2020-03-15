package gtk_history

import (
	"bytes"
	"log"
	"text/template"
)

var tmpl *template.Template

func init() {
	tmpl = new(template.Template)
	tmpl = template.Must(tmpl.Parse(
		`<span size="large">{{.Description}} {{ if .Project}} <span 
font_weight="bold" foreground="{{.Color}}">• {{.Project}}</span>{{ if .Client }} <span
foreground="#AAAAAA">• {{.Client}}</span>{{ end }}{{ end }}</span>
{{ if .Tag }}{{.Tag}} {{ end }}<span foreground="#AAAAAA" gravity="east">{{.Start}}{{ if .Stop }} - {{.Stop}}{{ else }}...{{ end }}</span>`))
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

func Render(c Ctx) string {
	buffer := bytes.Buffer{}
	if err := tmpl.Execute(&buffer, c); err != nil {
		log.Fatal(err)
	}
	return buffer.String()
}
