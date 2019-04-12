package widgets

import (
	"time"

	"github.com/burik666/yagostatus/ygs"
)

// ClockWidgetParams are widget parameters.
type ClockWidgetParams struct {
	Interval uint
	Format   string
}

// ClockWidget implements a clock.
type ClockWidget struct {
	params ClockWidgetParams
}

func init() {
	ygs.RegisterWidget("clock", NewClockWidget, ClockWidgetParams{
		Interval: 1,
		Format:   "Jan _2 Mon 15:04:05",
	})
}

// NewClockWidget returns a new ClockWidget.
func NewClockWidget(params interface{}) (ygs.Widget, error) {
	w := &ClockWidget{
		params: params.(ClockWidgetParams),
	}
	return w, nil
}

// Run starts the main loop.
func (w *ClockWidget) Run(c chan<- []ygs.I3BarBlock) error {
	ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
	res := []ygs.I3BarBlock{
		ygs.I3BarBlock{},
	}
	res[0].FullText = time.Now().Format(w.params.Format)
	c <- res
	for {
		select {
		case t := <-ticker.C:
			res[0].FullText = t.Format(w.params.Format)
			c <- res
		}
	}
}

// Event processes the widget events.
func (w *ClockWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) {}

// Stop shutdowns the widget.
func (w *ClockWidget) Stop() {}
