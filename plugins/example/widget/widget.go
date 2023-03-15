package widget

import (
	"github.com/burik666/yagostatus/ygs"
)

// Params are widget parameters.
type Params struct {
	Message string
}

// Widget implements a widget.
type Widget struct {
	ygs.BlankWidget

	params Params
}

// NewWidget returns a new Widget.
func NewWidget(params interface{}, wlogger ygs.Logger) (ygs.Widget, error) {
	w := &Widget{
		params: params.(Params),
	}

	return w, nil
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	c <- []ygs.I3BarBlock{
		{
			FullText: w.params.Message,
		},
	}

	return nil
}
