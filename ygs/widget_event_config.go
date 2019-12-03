package ygs

import (
	"fmt"
	"strings"
)

// WidgetEventConfig represents a widget events.
type WidgetEventConfig struct {
	Command      string   `yaml:"command"`
	Button       uint8    `yaml:"button"`
	Modifiers    []string `yaml:"modifiers,omitempty"`
	Name         string   `yaml:"name,omitempty"`
	Instance     string   `yaml:"instance,omitempty"`
	OutputFormat string   `yaml:"output_format,omitempty"`
	Override     bool     `yaml:"override"`
	WorkDir      string   `yaml:"workdir"`
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
		e.OutputFormat = "none"
	}

	return nil
}
