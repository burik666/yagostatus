package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/burik666/yagostatus/ygs"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config represents the main configuration.
type Config struct {
	Widgets []ConfigWidget
}

// ConfigWidget represents a widget configuration.
type ConfigWidget struct {
	Name     string
	Params   map[string]interface{}
	Template ygs.I3BarBlock
	Events   []ConfigWidgetEvent
}

// Validate checks widget configuration.
func (c ConfigWidget) Validate() error {
	if c.Name == "" {
		return errors.New("Missing widget name")
	}
	for _, e := range c.Events {
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ConfigWidgetEvent represents a widget events.
type ConfigWidgetEvent struct {
	Command   string   `yaml:"command"`
	Button    uint8    `yaml:"button"`
	Modifiers []string `yaml:"modifiers,omitempty"`
	Name      string   `yaml:"name,omitempty"`
	Instance  string   `yaml:"instance,omitempty"`
	Output    bool     `yaml:"output,omitempty"`
}

// Validate checks event parameters.
func (e ConfigWidgetEvent) Validate() error {
	var availableWidgetEventModifiers = [...]string{"Shift", "Control", "Mod1", "Mod2", "Mod3", "Mod4", "Mod5"}
	for _, mod := range e.Modifiers {
		found := false
		mod = strings.TrimLeft(mod, "!")
		for _, m := range availableWidgetEventModifiers {
			if mod == m {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("Unknown '%s' modifier", mod)
		}
	}
	return nil
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseConfig(data)
}

func parseConfig(data []byte) (*Config, error) {

	var tmp struct {
		Widgets []map[string]interface{} `yaml:"widgets"`
	}
	if err := yaml.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}

	config := Config{}

	for _, v := range tmp.Widgets {
		widget := ConfigWidget{}

		widget.Name, _ = v["widget"].(string)
		delete(v, "widget")

		tpl, ok := v["template"]
		if ok {
			if err := json.Unmarshal([]byte(tpl.(string)), &widget.Template); err != nil {
				return nil, err
			}
			delete(v, "template")
		}

		events, ok := v["events"]
		if ok {
			ymlevents, _ := yaml.Marshal(events)
			yaml.Unmarshal(ymlevents, &widget.Events)
			delete(v, "events")
		}

		widget.Params = v
		if err := widget.Validate(); err != nil {
			return nil, err
		}
		config.Widgets = append(config.Widgets, widget)
	}
	return &config, nil
}
