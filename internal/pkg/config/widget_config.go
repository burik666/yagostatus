package config

import (
	"errors"
	"fmt"

	"github.com/burik666/yagostatus/ygs"
)

// WidgetConfig represents a widget configuration.
type WidgetConfig struct {
	Name       string              `yaml:"widget"`
	Workspaces []string            `yaml:"workspaces"`
	Templates  []ygs.I3BarBlock    `yaml:"-"`
	Events     []WidgetEventConfig `yaml:"events"`
	WorkDir    string              `yaml:"-"`
	Index      int                 `yaml:"-"`
	File       string              `yaml:"-"`
	Variables  map[string]string   `yaml:"variables"`

	Params map[string]interface{} `yaml:",inline"`

	IncludeStack []string `yaml:"-"`
}

// Validate checks widget configuration.
func (c WidgetConfig) Validate() error {
	if c.Name == "" {
		return errors.New("missing widget name")
	}

	for ei := range c.Events {
		if err := c.Events[ei].Validate(); err != nil {
			return fmt.Errorf("events#%d: %s", ei+1, err)
		}
	}

	return nil
}
