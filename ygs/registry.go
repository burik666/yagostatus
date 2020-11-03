package ygs

import (
	"fmt"
	"reflect"

	rs "github.com/burik666/yagostatus/internal/registry/store"
)

// WidgetSpec describes constructor for widgets.
type WidgetSpec struct {
	Name          string
	NewFunc       NewWidgetFunc
	DefaultParams interface{}
}

// NewWidgetFunc function to create a new instance of a widget.
type NewWidgetFunc = func(params interface{}, l Logger) (Widget, error)

// RegisterWidget registers widget.
func RegisterWidget(rw WidgetSpec) error {
	def := reflect.ValueOf(rw.DefaultParams)
	if def.Kind() != reflect.Struct {
		return fmt.Errorf("defaultParams should be a struct")
	}

	if _, loaded := rs.LoadOrStore(rw.Name, rw); loaded {
		return fmt.Errorf("widget '%s' already registered", rw.Name)
	}

	return nil
}

// UnregisterWidget unregisters widget.
func UnregisterWidget(name string) bool {
	_, ok := rs.LoadAndDelete(name)

	return ok
}

// RegisteredWidgets returns list of registered widgets.
func RegisteredWidgets() []WidgetSpec {
	var widgets []WidgetSpec

	rs.Range(func(k, v interface{}) bool {
		widgets = append(widgets, v.(WidgetSpec))

		return true
	})

	return widgets
}
