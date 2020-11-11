package plugin

import (
	"github.com/burik666/yagostatus/plugins/example/widget"
	"github.com/burik666/yagostatus/ygs"
)

// Params contains example plugin parameters.
type Params struct {
	DefaultMessage string `yaml:"default_message"`
}

var Spec = ygs.PluginSpec{
	Name: "example",
	DefaultParams: Params{
		"not set",
	},
	InitFunc: func(p interface{}, l ygs.Logger) error {
		params := p.(Params)
		l.Infof("params: %+v", params)

		if err := ygs.RegisterWidget(ygs.WidgetSpec{
			Name:    "example",
			NewFunc: widget.NewWidget,
			DefaultParams: widget.Params{
				Message: params.DefaultMessage,
			},
		}); err != nil {
			panic(err)
		}

		return nil
	},

	ShutdownFunc: func() error {
		return nil
	},
}
