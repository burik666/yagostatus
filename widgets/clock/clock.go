package clock

import (
	"github.com/burik666/yagostatus/widgets/blank"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
)

// ClockWidgetParams are widget parameters.
type WidgetParams struct {
	Interval uint
	Format   string
}

// ClockWidget implements a clock.
type Widget struct {
	blank.Widget
	params WidgetParams
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("clock", NewClockWidget, WidgetParams{
		Interval: 1,
		Format:   "Jan _2 Mon 15:04:05",
	})
}

// NewClockWidget returns a new ClockWidget.
func NewClockWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	w := &Widget{
		params: params.(WidgetParams),
	}

	return w, nil
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	res := []ygs.I3BarBlock{
		{},
	}
	res[0].FullText = time.Now().Format(w.params.Format)

	c <- res

	ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
	for t := range ticker.C {
		res[0].FullText = t.Format(w.params.Format)
		c <- res
	}

	return nil
}
