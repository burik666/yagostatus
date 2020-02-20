package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/burik666/yagostatus/ygs"
)

// Config represents the main configuration.
type Config struct {
	Signals struct {
		StopSignal syscall.Signal `yaml:"stop"`
		ContSignal syscall.Signal `yaml:"cont"`
	} `yaml:"signals"`
	Widgets []ygs.WidgetConfig `yaml:"widgets"`
}

// LoadFile loads and parses config from file.
func LoadFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return parse(data, filepath.Dir(filename), filepath.Base(filename))
}

// Parse parses config.
func Parse(data []byte, source string) (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return parse(data, wd, source)
}
