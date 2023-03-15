package ygs

import (
	"fmt"
	"reflect"
	"strings"

	rs "github.com/burik666/yagostatus/internal/registry/store"
)

// WidgetSpec describes constructor for widgets.
type WidgetSpec struct {
	Name          string
	DefaultParams interface{}
	NewFunc       NewWidgetFunc
}

// WidgetSpec describes plugins initialization.
type PluginSpec struct {
	Name          string
	DefaultParams interface{}
	InitFunc      func(params interface{}, l Logger) error
	ShutdownFunc  func() error
}

// NewWidgetFunc function to create a new instance of a widget.
type NewWidgetFunc = func(params interface{}, l Logger) (Widget, error)

// RegisterWidget registers widget.
func RegisterWidget(rw WidgetSpec) error {
	if rw.DefaultParams != nil {
		def := reflect.ValueOf(rw.DefaultParams)
		if def.Kind() != reflect.Struct {
			return fmt.Errorf("defaultParams should be a struct")
		}
	}

	if _, loaded := rs.LoadOrStore("widget_"+rw.Name, rw); loaded {
		return fmt.Errorf("widget '%s' already registered", rw.Name)
	}

	return nil
}

// UnregisterWidget unregisters widget.
func UnregisterWidget(name string) bool {
	_, ok := rs.LoadAndDelete("widget_" + name)

	return ok
}

// RegisteredWidgets returns list of registered plugins.
func RegisteredWidgets() []WidgetSpec {
	var widgets []WidgetSpec

	rs.Range(func(k, v interface{}) bool {
		if strings.HasPrefix(k.(string), "widget_") {
			widgets = append(widgets, v.(WidgetSpec))
		}

		return true
	})

	return widgets
}

// RegisterPlugin registers plugin.
func RegisterPlugin(rw PluginSpec) error {
	if rw.DefaultParams != nil {
		def := reflect.ValueOf(rw.DefaultParams)
		if def.Kind() != reflect.Struct {
			return fmt.Errorf("defaultParams should be a struct")
		}
	}

	if _, loaded := rs.LoadOrStore("plugin_"+rw.Name, rw); loaded {
		return fmt.Errorf("plugin '%s' already registered", rw.Name)
	}

	return nil
}

// UnregisterPlugin unregisters widget.
func UnregisterPlugin(name string) bool {
	_, ok := rs.LoadAndDelete("plugin_" + name)

	return ok
}

// RegisteredPlugins returns list of registered plugins.
func RegisteredPlugins() []PluginSpec {
	var plugins []PluginSpec

	rs.Range(func(k, v interface{}) bool {
		if strings.HasPrefix(k.(string), "plugin_") {
			plugins = append(plugins, v.(PluginSpec))
		}

		return true
	})

	return plugins
}
