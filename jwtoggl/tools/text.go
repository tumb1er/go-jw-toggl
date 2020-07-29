package tools

const TrayLen = 20

func Shorten(desc string) string {
	runes := []rune(desc)
	l := len(runes)
	if l > TrayLen {
		runes = append(runes[:TrayLen], []rune("...")...)
	}
	return string(runes)
}
