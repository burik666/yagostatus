package ygs

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

type newWidgetFunc = func(interface{}) (Widget, error)

type widget struct {
	newFunc       newWidgetFunc
	defaultParams interface{}
}

var registeredWidgets = make(map[string]widget)

// RegisterWidget registers widget.
func RegisterWidget(name string, newFunc newWidgetFunc, defaultParams interface{}) {
	if _, ok := registeredWidgets[name]; ok {
		panic(fmt.Sprintf("Widget '%s' already registered", name))
	}
	def := reflect.ValueOf(defaultParams)
	if def.Kind() != reflect.Struct {
		panic("defaultParams should be a struct")
	}
	registeredWidgets[name] = widget{
		newFunc:       newFunc,
		defaultParams: defaultParams,
	}
}

// NewWidget creates new widget by name.
func NewWidget(name string, rawParams map[string]interface{}) (Widget, error) {
	widget, ok := registeredWidgets[name]
	if !ok {
		return nil, fmt.Errorf("Widget '%s' not found", name)
	}

	def := reflect.ValueOf(widget.defaultParams)

	params := reflect.New(def.Type())
	params.Elem().Set(def)

	b, _ := yaml.Marshal(rawParams)
	if err := yaml.UnmarshalStrict(b, params.Interface()); err != nil {
		return nil, err
	}

	return widget.newFunc(params.Elem().Interface())
}

// ErrorWidget creates new widget with error message.
func ErrorWidget(text string) (string, map[string]interface{}) {
	blocks, _ := json.Marshal([]I3BarBlock{
		I3BarBlock{
			FullText: text,
			Color:    "#ff0000",
		},
	})

	return "static", map[string]interface{}{
		"blocks": string(blocks),
	}
}
