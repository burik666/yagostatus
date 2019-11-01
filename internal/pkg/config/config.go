package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"syscall"

	"github.com/burik666/yagostatus/internal/pkg/executor"
	"github.com/burik666/yagostatus/ygs"

	"gopkg.in/yaml.v2"
)

// Config represents the main configuration.
type Config struct {
	Signals struct {
		StopSignal syscall.Signal `yaml:"stop"`
		ContSignal syscall.Signal `yaml:"cont"`
	} `yaml:"signals"`
	Widgets []WidgetConfig `yaml:"widgets"`
}

// WidgetConfig represents a widget configuration.
type WidgetConfig struct {
	Name       string              `yaml:"widget"`
	Workspaces []string            `yaml:"workspaces"`
	Template   ygs.I3BarBlock      `yaml:"-"`
	Events     []WidgetEventConfig `yaml:"events"`

	Params map[string]interface{}
}

// Validate checks widget configuration.
func (c WidgetConfig) Validate() error {
	if c.Name == "" {
		return errors.New("missing widget name")
	}

	for ei := range c.Events {
		if err := c.Events[ei].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// WidgetEventConfig represents a widget events.
type WidgetEventConfig struct {
	Command      string                `yaml:"command"`
	Button       uint8                 `yaml:"button"`
	Modifiers    []string              `yaml:"modifiers,omitempty"`
	Name         string                `yaml:"name,omitempty"`
	Instance     string                `yaml:"instance,omitempty"`
	OutputFormat executor.OutputFormat `yaml:"output_format,omitempty"`
}

// Validate checks event parameters.
func (e *WidgetEventConfig) Validate() error {
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
			return fmt.Errorf("unknown '%s' modifier", mod)
		}
	}

	if e.OutputFormat == "" {
		e.OutputFormat = executor.OutputFormatNone
	}

	return nil
}

// LoadFile loads and parses config from file.
func LoadFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

// Parse parses config.
func Parse(data []byte) (*Config, error) {
	var raw struct {
		Widgets []map[string]interface{} `yaml:"widgets"`
	}

	config := Config{}
	config.Signals.StopSignal = syscall.SIGUSR1
	config.Signals.ContSignal = syscall.SIGCONT

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, trimYamlErr(err, false)
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, trimYamlErr(err, false)
	}

	for widgetIndex := range config.Widgets {
		widget := &config.Widgets[widgetIndex]
		params := raw.Widgets[widgetIndex]

		tpl, ok := params["template"]
		if ok {
			if err := json.Unmarshal([]byte(tpl.(string)), &widget.Template); err != nil {
				name, params := ygs.ErrorWidget(err.Error())
				*widget = WidgetConfig{
					Name:   name,
					Params: params,
				}

				log.Printf("template error: %s", err)

				continue
			}
		}

		tmp, err := yaml.Marshal(params["events"])
		if err != nil {
			return nil, err
		}

		if err := yaml.UnmarshalStrict(tmp, &widget.Events); err != nil {
			name, params := ygs.ErrorWidget(trimYamlErr(err, true).Error())
			*widget = WidgetConfig{
				Name:   name,
				Params: params,
			}

			continue
		}

		widget.Params = params
		if err := widget.Validate(); err != nil {
			name, params := ygs.ErrorWidget(trimYamlErr(err, true).Error())
			*widget = WidgetConfig{
				Name:   name,
				Params: params,
			}

			continue
		}

		delete(params, "widget")
		delete(params, "workspaces")
		delete(params, "template")
		delete(params, "events")
	}

	return &config, nil
}

func trimYamlErr(err error, trimLineN bool) error {
	msg := strings.TrimPrefix(err.Error(), "yaml: unmarshal errors:\n  ")
	if trimLineN {
		msg = strings.TrimPrefix(msg, "line ")
		msg = strings.TrimLeft(msg, "1234567890: ")
	}

	return errors.New(msg)
}
