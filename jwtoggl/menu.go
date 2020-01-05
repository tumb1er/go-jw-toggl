package jwtoggl

func (j *jApp) initMenu(b *builder) {
	b.builder.ConnectSignals(map[string]interface{}{
		"toggle": j.onToggle,
		"exit": j.onExit,
	})
}