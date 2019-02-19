package widgets

import (
	"errors"
	"time"

	"github.com/burik666/yagostatus/ygs"
)

// ClockWidget implements a clock.
type ClockWidget struct {
	format   string
	interval time.Duration
}

// NewClockWidget returns a new ClockWidget.
func NewClockWidget(params map[string]interface{}) (ygs.Widget, error) {
	w := &ClockWidget{}

	v, ok := params["format"]
	if !ok {
		return nil, errors.New("missing 'format' setting")
	}
	w.format = v.(string)

	v, ok = params["interval"]
	if ok {
		w.interval = time.Duration(v.(int)) * time.Second
	} else {
		w.interval = time.Second
	}
	return w, nil
}

// Run starts the main loop.
func (w *ClockWidget) Run(c chan<- []ygs.I3BarBlock) error {
	ticker := time.NewTicker(w.interval)
	res := []ygs.I3BarBlock{
		ygs.I3BarBlock{},
	}
	res[0].FullText = time.Now().Format(w.format)
	c <- res
	for {
		select {
		case t := <-ticker.C:
			res[0].FullText = t.Format(w.format)
			c <- res
		}
	}
}

// Event processes the widget events.
func (w *ClockWidget) Event(event ygs.I3BarClickEvent) {}

// Stop shutdowns the widget.
func (w *ClockWidget) Stop() {}

func init() {
	ygs.RegisterWidget("clock", NewClockWidget)
}
