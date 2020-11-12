package registry

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/burik666/yagostatus/internal/config"
	rs "github.com/burik666/yagostatus/internal/registry/store"
	"github.com/burik666/yagostatus/ygs"
	"gopkg.in/yaml.v2"
)

// NewWidget creates new widget by name.
func NewWidget(widgetConfig config.WidgetConfig, wlogger ygs.Logger) (ygs.Widget, error) {
	name := widgetConfig.Name
	wi, ok := rs.Load("widget_" + name)

	if !ok {
		return nil, fmt.Errorf("widget '%s' not found", name)
	}

	widget := wi.(ygs.WidgetSpec)
	if widget.DefaultParams == nil {
		return widget.NewFunc(nil, wlogger)
	}

	def := reflect.ValueOf(widget.DefaultParams)

	params := reflect.New(def.Type())
	pe := params.Elem()
	pe.Set(def)

	delete(widgetConfig.Params, "template")
	delete(widgetConfig.Params, "templates")

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

	return widget.NewFunc(pe.Interface(), wlogger)
}

func trimYamlErr(err error, trimLineN bool) error {
	msg := strings.TrimPrefix(err.Error(), "yaml: unmarshal errors:\n  ")
	if trimLineN {
		msg = strings.TrimPrefix(msg, "line ")
		msg = strings.TrimLeft(msg, "1234567890: ")
	}

	return errors.New(msg)
}
