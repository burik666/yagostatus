package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// Config represents the main configuration.
type Config struct {
	Signals struct {
		StopSignal syscall.Signal `yaml:"stop"`
		ContSignal syscall.Signal `yaml:"cont"`
	} `yaml:"signals"`
	Plugins struct {
		Path string   `yaml:"path"`
		Load []string `yaml:"load"`
	} `yaml:"plugins"`
	Variables map[string]interface{} `yaml:"variables"`
	Widgets   []WidgetConfig         `yaml:"widgets"`
	File      string                 `yaml:"-"`
}

// SnippetConfig represents the snippet configuration.
type SnippetConfig struct {
	Variables map[string]interface{} `yaml:"variables"`
	Widgets   []WidgetConfig         `yaml:"widgets"`
}

// LoadFile loads and parses config from file.
func LoadFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg, err := parse(data, filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return nil, err
	}

	cfg.File = filename

	return cfg, nil
}

// Parse parses config.
func Parse(data []byte, source string) (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg, err := parse(data, wd, source)
	if err != nil {
		return nil, err
	}

	cfg.File = source

	return cfg, nil
}
