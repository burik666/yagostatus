package main

import (
	"encoding/json"
	"errors"
	"github.com/burik666/yagostatus/ygs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type tmpConfig struct {
	Widgets []map[string]interface{} `yaml:"widgets"`
}

type Config struct {
	Widgets []ConfigWidget
}

type ConfigWidgetEvent struct {
	Command  string `yaml:"command"`
	Button   uint8  `yaml:"button"`
	Name     string `yaml:"name,omitempty"`
	Instance string `yaml:"instance,omitempty"`
	Output   bool   `yaml:"output,omitempty"`
}

type ConfigWidget struct {
	Name     string
	Params   map[string]interface{}
	Template ygs.I3BarBlock
	Events   []ConfigWidgetEvent
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var t tmpConfig
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	config := Config{}

	for _, v := range t.Widgets {
		widget := ConfigWidget{}

		var ok bool
		widget.Name, ok = v["widget"].(string)
		if !ok {
			return nil, errors.New("Missing widget name.")
		}
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
		config.Widgets = append(config.Widgets, widget)
	}
	return &config, nil
}
