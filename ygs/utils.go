package ygs

import (
	"log"
	"reflect"
	"strings"
)

var registeredWidgets = make(map[string]reflect.Type)

// RegisterWidget registers widget.
func RegisterWidget(widget Widget) {
	t := reflect.TypeOf(widget).Elem()
	name := strings.ToLower(t.Name())
	if _, ok := registeredWidgets[name]; ok {
		log.Fatalf("Widget '%s' already registered", name)
	}
	registeredWidgets[name] = t
}

// NewWidget creates new widget by name.
func NewWidget(name string) (Widget, bool) {
	t, ok := registeredWidgets[name]
	if !ok {
		return nil, false
	}
	v := reflect.New(t)
	return v.Interface().(Widget), true
}
