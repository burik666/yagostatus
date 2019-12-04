package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
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

		for i := range widget.Events {
			if widget.Events[i].WorkDir == "" {
				widget.Events[i].WorkDir = workdir
			}
		}

		// for backward compatibility
		if itpl, ok := params["template"]; ok {
			tpl, ok := itpl.(string)
			if !ok {
				setError(widget, fmt.Errorf("invalid template"), false)
				continue WIDGET
			}

			widget.Templates = append(widget.Templates, ygs.I3BarBlock{})
			if err := json.Unmarshal([]byte(tpl), &widget.Templates[0]); err != nil {
				setError(widget, err, false)

				continue WIDGET
			}

			delete(params, "template")
		}

		if itpls, ok := params["templates"]; ok {
			tpls, ok := itpls.(string)
			if !ok {
				setError(widget, fmt.Errorf("invalid templates"), false)
				continue WIDGET
			}

			if err := json.Unmarshal([]byte(tpls), &widget.Templates); err != nil {
				setError(widget, err, false)

				continue WIDGET
			}

			delete(params, "templates")
		}

		tpls, _ := json.Marshal(widget.Templates)

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

			filename := widget.Name[1:]
			if !filepath.IsAbs(filename) {
				filename = wd + "/" + filename
			}

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
				json.Unmarshal(tpls, &snipWidgetsConfig[i].Templates)

				snipEvents := snipWidgetsConfig[i].Events
				for i := range snipEvents {
					if snipEvents[i].WorkDir == "" {
						snipEvents[i].WorkDir = wd
					}
				}

				for _, e := range widget.Events {
					if e.Override {
						sort.Strings(e.Modifiers)

						ne := make([]ygs.WidgetEventConfig, 0, len(snipEvents))

						for _, se := range snipEvents {
							sort.Strings(se.Modifiers)

							if e.Button == se.Button &&
								e.Name == se.Name &&
								e.Instance == se.Instance &&
								reflect.DeepEqual(e.Modifiers, se.Modifiers) {

								continue
							}

							ne = append(ne, se)
						}
						snipEvents = append(ne, e)
					} else {
						snipEvents = append(snipEvents, e)
					}
				}

				snipWidgetsConfig[i].Events = snipEvents
			}

			i := widgetIndex
			config.Widgets = append(config.Widgets[:i], config.Widgets[i+1:]...)
			config.Widgets = append(config.Widgets[:i], append(snipWidgetsConfig, config.Widgets[i:]...)...)

			widgetIndex--

			continue WIDGET
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

		if n, err := strconv.ParseInt(st, 10, 64); err == nil {
			vi := reflect.New(reflect.ValueOf(n).Type()).Elem()
			vi.SetInt(n)
			*v = vi
			return
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
