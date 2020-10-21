package ygs

import (
	"fmt"
	"reflect"

	rs "github.com/burik666/yagostatus/internal/pkg/registry/store"
)

type RegistryWidget struct {
	Name          string
	NewFunc       NewWidgetFunc
	DefaultParams interface{}
}

// RegisterWidget registers widget.
func RegisterWidget(name string, newFunc NewWidgetFunc, defaultParams interface{}) error {
	def := reflect.ValueOf(defaultParams)
	if def.Kind() != reflect.Struct {
		return fmt.Errorf("defaultParams should be a struct")
	}

	if _, loaded := rs.LoadOrStore(name, RegistryWidget{
		Name:          name,
		NewFunc:       newFunc,
		DefaultParams: defaultParams,
	}); loaded {
		return fmt.Errorf("widget '%s' already registered", name)
	}

	return nil
}

func UnregisterWidget(name string) bool {
	_, ok := rs.LoadAndDelete(name)

	return ok
}

func RegisteredWidgets() []RegistryWidget {
	var widgets []RegistryWidget

	rs.Range(func(k, v interface{}) bool {
		widgets = append(widgets, v.(RegistryWidget))

		return true
	})

	return widgets
}
