package tools

const TrayLen = 20

func Shorten(desc string) string {
	l := len(desc)
	if l > TrayLen {
		desc = desc[:TrayLen] + "..."
	}
	return desc
}
