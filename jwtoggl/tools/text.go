package tools

func Shorten(desc string) string {
	l := len(desc)
	if l > 10 {
		desc = " " + desc[:10] + "..."
	} else {
		desc = " " + desc
	}
	return desc
}
