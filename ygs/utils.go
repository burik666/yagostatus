package ygs

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"gopkg.in/yaml.v2"
)

type newWidgetFunc = func(interface{}, logger.Logger) (Widget, error)

type widget struct {
	newFunc       newWidgetFunc
	defaultParams interface{}
}

var registeredWidgets = make(map[string]widget)

// RegisterWidget registers widget.
func RegisterWidget(name string, newFunc newWidgetFunc, defaultParams interface{}) {
	if _, ok := registeredWidgets[name]; ok {
		panic(fmt.Sprintf("widget '%s' already registered", name))
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
func NewWidget(widgetConfig WidgetConfig, wlogger logger.Logger) (Widget, error) {
	name := widgetConfig.Name
	widget, ok := registeredWidgets[name]
	if !ok {
		return nil, fmt.Errorf("widget '%s' not found", name)
	}

	def := reflect.ValueOf(widget.defaultParams)

	params := reflect.New(def.Type())
	pe := params.Elem()
	pe.Set(def)

	b, err := yaml.Marshal(widgetConfig.Params)
	if err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(b, params.Interface()); err != nil {
		return nil, trimYamlErr(err, true)
	}

	if _, ok := widgetConfig.Params["workdir"]; !ok {
		for i := 0; i < pe.NumField(); i++ {
			fn := pe.Type().Field(i).Name
			if strings.ToLower(fn) == "workdir" {
				pe.Field(i).SetString(widgetConfig.WorkDir)
			}
		}
	}

	return widget.newFunc(pe.Interface(), wlogger)
}

// ErrorWidget creates new widget with error message.
func ErrorWidget(text string) WidgetConfig {
	blocks, _ := json.Marshal([]I3BarBlock{
		{
			FullText: text,
			Color:    "#ff0000",
		},
	})

	return WidgetConfig{
		Name: "static",
		Params: map[string]interface{}{
			"blocks": string(blocks),
		},
		File: "bultin",
	}

}

func trimYamlErr(err error, trimLineN bool) error {
	msg := strings.TrimPrefix(err.Error(), "yaml: unmarshal errors:\n  ")
	if trimLineN {
		msg = strings.TrimPrefix(msg, "line ")
		msg = strings.TrimLeft(msg, "1234567890: ")
	}

	return errors.New(msg)
}
