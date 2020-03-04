package tools

import (
	"github.com/faiface/mainthread"
	"sync"
)

type UI struct {
	mu         sync.Mutex
	mainThread bool
}

type ParamValueErrorCallable func(interface{}) (interface{}, error)
type ValueErrorCallable func() (interface{}, error)
type ParamCallable func(interface{})
type ErrorCallable func() error

func (ui *UI) inMainThread() bool {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	return ui.mainThread
}

func (ui *UI) setMainThread(value bool) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.mainThread = value
}

func (ui *UI) CallParamValueError(c ParamValueErrorCallable, p interface{}) (interface{}, error) {
	if ui.inMainThread() {
		return c(p)
	} else {
		ui.setMainThread(true)
		defer ui.setMainThread(false)
		var result interface{}
		var err error
		mainthread.Call(func() {
			result, err = c(p)
		})
		return result, err
	}
}

func (ui *UI) CallValueError(c ValueErrorCallable) (interface{}, error) {
	return ui.CallParamValueError(func(_ interface{}) (interface{}, error) {
		return c()
	}, nil)
}

func (ui *UI) CallParam(c ParamCallable, p interface{}) {
	_, _ = ui.CallParamValueError(func(_ interface{}) (interface{}, error) {
		c(p)
		return nil, nil
	}, p)
}

func (ui *UI) CallError(c ErrorCallable) error {
	_, err := ui.CallParamValueError(func(_ interface{}) (interface{}, error) {
		return nil, c()
	}, nil)
	return err
}
