package config

import (
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"

	"github.com/burik666/yagostatus/ygs"
	"gopkg.in/yaml.v2"
)

type PluginConfig struct {
	Plugin string                 `yaml:"plugin"`
	Params map[string]interface{} `yaml:",inline"`
}

func LoadPlugins(cfg Config, logger ygs.Logger) error {
	for _, l := range cfg.Plugins.Load {
		fname := l.Plugin

		if !filepath.IsAbs(fname) {
			fname = filepath.Join(cfg.Plugins.Path, fname)
		}

		logger.Infof("Load plugin: %s", fname)

		p, err := plugin.Open(fname)
		if err != nil {
			return err
		}

		plugin, err := p.Lookup("Plugin")
		if err != nil {
			return fmt.Errorf("variable Plugin: %w", err)
		}

		pv := reflect.ValueOf(plugin)

		specp, ok := pv.Interface().(*ygs.PluginSpec)
		if !ok {
			return fmt.Errorf("variable Plugin is not a ygs.PluginSpec")
		}

		spec := (*specp)

		spec.Name = fmt.Sprintf("%s#%s", l.Plugin, spec.Name)

		if spec.DefaultParams != nil {
			pb, err := yaml.Marshal(l.Params)
			if err != nil {
				return err
			}

			params := reflect.New(reflect.TypeOf(spec.DefaultParams))
			params.Elem().Set(reflect.ValueOf(spec.DefaultParams))

			if err := yaml.UnmarshalStrict(pb, params.Interface()); err != nil {
				return trimYamlErr(err, true)
			}

			spec.DefaultParams = params.Elem().Interface()
		}

		if err := ygs.RegisterPlugin(spec); err != nil {
			return fmt.Errorf("failed to register plugin: %w", err)
		}
	}

	return nil
}

func InitPlugins(logger ygs.Logger) error {
	for _, yp := range ygs.RegisteredPlugins() {
		if yp.InitFunc != nil {
			logger.Infof("Init plugin: %s", yp.Name)

			l := logger.WithPrefix(fmt.Sprintf("[%s]", yp.Name))
			if err := yp.InitFunc(yp.DefaultParams, l); err != nil {
				logger.Errorf("init [%s]: %s", yp.Name, err)
			}
		}
	}

	return nil
}

func ShutdownPlugins(logger ygs.Logger) {
	for _, yp := range ygs.RegisteredPlugins() {
		if yp.ShutdownFunc != nil {
			logger.Infof("Shutdown plugin: %s", yp.Name)

			if err := yp.ShutdownFunc(); err != nil {
				logger.Errorf("shutdown [%s]: %s", yp.Name, err)
			}
		}
	}
}
