package tools

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
)

func GetString(model *gtk.TreeModel, iter *gtk.TreeIter, column int) (string, error) {
	glibValue, err := model.GetValue(iter, column)
	if err != nil {
		return "", err
	}
	v, err := glibValue.GoValue()
	if err != nil {
		return "", err
	}
	result, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("string convertion error: %T", v)
	}
	return result, nil
}
