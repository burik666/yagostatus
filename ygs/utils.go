package ygs

import (
	"log"

	"github.com/pkg/errors"
)

type newWidgetFunc = func(map[string]interface{}) (Widget, error)

var registeredWidgets = make(map[string]newWidgetFunc)

// RegisterWidget registers widget.
func RegisterWidget(name string, newFunc newWidgetFunc) {
	if _, ok := registeredWidgets[name]; ok {
		log.Fatalf("Widget '%s' already registered", name)
	}
	registeredWidgets[name] = newFunc
}

// NewWidget creates new widget by name.
func NewWidget(name string, params widgetParams) (Widget, error) {
	newFunc, ok := registeredWidgets[name]
	if !ok {
		return nil, errors.Errorf("Widget '%s' not found", name)
	}
	return newFunc(params)
}
