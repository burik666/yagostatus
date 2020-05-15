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

func parse(data []byte, workdir string, source string) (*Config, error) {

	config := Config{}
	config.Signals.StopSignal = syscall.SIGUSR1
	config.Signals.ContSignal = syscall.SIGCONT

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, trimYamlErr(err, false)
	}

	for wi := range config.Widgets {
		config.Widgets[wi].File = source
		config.Widgets[wi].Index = wi
	}

	dict := make(map[string]string, len(config.Variables))
	for k, v := range config.Variables {
		vb, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		var vraw ygs.Vary

		err = json.Unmarshal(vb, &vraw)
		if err != nil {
			return nil, err
		}

		dict[fmt.Sprintf("${%s}", k)] = strings.TrimRight(vraw.String(), "\n")
	}

	v := reflect.ValueOf(config.Widgets)
	replaceRecursive(&v, dict)

WIDGET:
	for wi := 0; wi < len(config.Widgets); wi++ {
		widget := &config.Widgets[wi]

		params := make(map[string]interface{})
		for k, v := range config.Widgets[wi].Params {
			params[strings.ToLower(k)] = v
		}

		config.Widgets[wi].Params = params

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

		ok, err := parseSnippet(&config, wi, params)
		if err != nil {
			setError(widget, err, false)

			continue WIDGET
		}

		if ok {
			wi--
			continue WIDGET
		}

		if err := widget.Validate(); err != nil {
			setError(widget, err, true)

			continue WIDGET
		}
	}

	return &config, nil
}

func parseSnippet(config *Config, wi int, params map[string]interface{}) (bool, error) {
	widget := config.Widgets[wi]

	if len(widget.Name) > 0 && widget.Name[0] == '$' {
		for i := range widget.IncludeStack {
			if widget.Name == widget.IncludeStack[i] {
				stack := append(widget.IncludeStack, widget.Name)

				return false, fmt.Errorf("recursive include: '%s'", strings.Join(stack, " -> "))
			}
		}

		wd := widget.WorkDir

		filename := widget.Name[1:]
		if !filepath.IsAbs(filename) {
			filename = wd + "/" + filename
		}

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return false, err
		}

		var snippetConfig SnippetConfig
		if err := yaml.Unmarshal(data, &snippetConfig); err != nil {
			return false, err
		}

		for k, v := range snippetConfig.Variables {
			if _, ok := params[k]; !ok {
				params[k] = v
			}
		}

		dict := make(map[string]string, len(params))
		for k, v := range params {
			if _, ok := snippetConfig.Variables[k]; !ok {
				return false, fmt.Errorf("unknown variable '%s'", k)
			}

			vb, err := json.Marshal(v)
			if err != nil {
				return false, err
			}

			var vraw ygs.Vary

			err = json.Unmarshal(vb, &vraw)
			if err != nil {
				return false, err
			}

			dict[fmt.Sprintf("${%s}", k)] = strings.TrimRight(vraw.String(), "\n")
		}

		v := reflect.ValueOf(snippetConfig.Widgets)
		replaceRecursive(&v, dict)

		tpls, _ := json.Marshal(widget.Templates)

		wd = filepath.Dir(filename)
		for i := range snippetConfig.Widgets {
			snippetConfig.Widgets[i].WorkDir = wd
			snippetConfig.Widgets[i].File = filename
			snippetConfig.Widgets[i].Index = i
			snippetConfig.Widgets[i].IncludeStack = append(widget.IncludeStack, widget.Name)
			json.Unmarshal(tpls, &snippetConfig.Widgets[i].Templates)

			snipEvents := snippetConfig.Widgets[i].Events
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

			snippetConfig.Widgets[i].Events = snipEvents
		}

		config.Widgets = append(config.Widgets[:wi], config.Widgets[wi+1:]...)
		config.Widgets = append(config.Widgets[:wi], append(snippetConfig.Widgets, config.Widgets[wi:]...)...)

		return true, nil
	}
	return false, nil
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
