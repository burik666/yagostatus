package ygs

import (
	"errors"
)

// WidgetConfig represents a widget configuration.
type WidgetConfig struct {
	Name       string              `yaml:"widget"`
	Workspaces []string            `yaml:"workspaces"`
	Templates  []I3BarBlock        `yaml:"-"`
	Events     []WidgetEventConfig `yaml:"events"`
	WorkDir    string              `yaml:"-"`

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
			return err
		}
	}

	return nil
}
