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
	ygs.BlankWidget

	params ClockWidgetParams
}

func init() {
	if err := ygs.RegisterWidget("clock", NewClockWidget, ClockWidgetParams{
		Interval: 1,
		Format:   "Jan _2 Mon 15:04:05",
	}); err != nil {
		panic(err)
	}
}

// NewClockWidget returns a new ClockWidget.
func NewClockWidget(params interface{}, wlogger ygs.Logger) (ygs.Widget, error) {
	w := &ClockWidget{
		params: params.(ClockWidgetParams),
	}

	return w, nil
}

// Run starts the main loop.
func (w *ClockWidget) Run(c chan<- []ygs.I3BarBlock) error {
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
