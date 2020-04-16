package models

import (
	"reflect"
)

type Named struct {
	Synced
	Name string
}

func (obj *Named) dump(out interface{}) error {
	if err := obj.Synced.dump(out); err != nil {
		return err
	}
	obj.Name = reflect.ValueOf(out).FieldByName("Name").String()
	return nil
}
