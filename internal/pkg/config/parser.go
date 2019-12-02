package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"

	"github.com/burik666/yagostatus/ygs"

	"gopkg.in/yaml.v2"
)

func parse(data []byte, workdir string) (*Config, error) {

	config := Config{}
	config.Signals.StopSignal = syscall.SIGUSR1
	config.Signals.ContSignal = syscall.SIGCONT

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, trimYamlErr(err, false)
	}

WIDGET:
	for widgetIndex := 0; widgetIndex < len(config.Widgets); widgetIndex++ {
		widget := &config.Widgets[widgetIndex]

		params := make(map[string]interface{})
		for k, v := range config.Widgets[widgetIndex].Params {
			params[strings.ToLower(k)] = v
		}

		config.Widgets[widgetIndex].Params = params

		if widget.WorkDir == "" {
			widget.WorkDir = workdir
		}

		if len(widget.Name) > 0 && widget.Name[0] == '$' {
			for i := range widget.IncludeStack {
				if widget.Name == widget.IncludeStack[i] {
					stack := append(widget.IncludeStack, widget.Name)

					setError(widget, fmt.Errorf("recursive include: '%s'", strings.Join(stack, " -> ")), false)

					continue WIDGET
				}
			}

			wd := workdir

			if widget.WorkDir != "" {
				wd = widget.WorkDir
			}

			filename := wd + "/" + widget.Name[1:]
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				setError(widget, err, false)

				continue WIDGET
			}

			dict := make(map[string]string, len(params))
			for k, v := range params {
				vb, err := json.Marshal(v)
				if err != nil {
					setError(widget, err, false)

					continue WIDGET
				}

				var vraw ygs.Vary

				err = json.Unmarshal(vb, &vraw)
				if err != nil {
					setError(widget, err, true)

					continue WIDGET
				}

				dict[fmt.Sprintf("${%s}", k)] = strings.TrimRight(vraw.String(), "\n")
			}

			var snipWidgetsConfig []ygs.WidgetConfig
			if err := yaml.Unmarshal(data, &snipWidgetsConfig); err != nil {
				setError(widget, err, false)

				continue WIDGET
			}

			v := reflect.ValueOf(snipWidgetsConfig)
			replaceRecursive(&v, dict)

			wd = filepath.Dir(filename)
			for i := range snipWidgetsConfig {
				snipWidgetsConfig[i].WorkDir = wd
				snipWidgetsConfig[i].IncludeStack = append(widget.IncludeStack, widget.Name)
			}

			i := widgetIndex
			config.Widgets = append(config.Widgets[:i], config.Widgets[i+1:]...)
			config.Widgets = append(config.Widgets[:i], append(snipWidgetsConfig, config.Widgets[i:]...)...)

			widgetIndex--

			continue WIDGET
		}

		if tpl, ok := params["template"]; ok {
			if err := json.Unmarshal([]byte(tpl.(string)), &widget.Template); err != nil {
				setError(widget, err, false)

				log.Printf("template error: %s", err)

				continue WIDGET
			}

			delete(params, "template")
		}

		if err := widget.Validate(); err != nil {
			setError(widget, err, true)

			continue WIDGET
		}
	}

	return &config, nil
}

func setError(widget *ygs.WidgetConfig, err error, trimLineN bool) {
	*widget = ygs.ErrorWidget(trimYamlErr(err, trimLineN).Error())
}

func trimYamlErr(err error, trimLineN bool) error {
	msg := strings.TrimPrefix(err.Error(), "yaml: ")
	msg = strings.TrimPrefix(msg, "unmarshal errors:\n  ")
	if trimLineN {
		msg = strings.TrimPrefix(msg, "line ")
		msg = strings.TrimLeft(msg, "1234567890: ")
	}

	return errors.New(msg)
}

func replaceRecursive(v *reflect.Value, dict map[string]string) {
	vv := *v
	for vv.Kind() == reflect.Ptr || vv.Kind() == reflect.Interface {
		vv = vv.Elem()
	}

	switch vv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < vv.Len(); i++ {
			vi := vv.Index(i)
			replaceRecursive(&vi, dict)
		}
	case reflect.Map:
		for _, i := range vv.MapKeys() {
			vm := vv.MapIndex(i)
			replaceRecursive(&vm, dict)
			vv.SetMapIndex(i, vm)
		}
	case reflect.Struct:
		t := vv.Type()
		for i := 0; i < t.NumField(); i++ {
			vf := v.Field(i)
			replaceRecursive(&vf, dict)
		}
	case reflect.String:
		st := vv.String()
		for s, r := range dict {
			st = strings.ReplaceAll(st, s, r)
		}

		if vv.CanSet() {
			vv.SetString(st)
		} else {
			vn := reflect.New(vv.Type()).Elem()
			vn.SetString(st)
			*v = vn
		}
	}
}
