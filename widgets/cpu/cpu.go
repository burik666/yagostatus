package clock

import (
	"fmt"
	"github.com/burik666/yagostatus/widgets/blank"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
	"github.com/shirou/gopsutil/cpu"
)

type WidgetParams struct {
	Interval uint
	logger   logger.Logger
	Format   string
}

type Widget struct {
	blank.Widget
	params WidgetParams
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("cpu", NewWidget, WidgetParams{
		Interval: 1,
	})
}

// NewWidget returns a new ClockWidget.
func NewWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	w := &Widget{
		params: params.(WidgetParams),
	}
	w.params.logger = wlogger

	return w, nil
}

func (w *Widget) getCPU() string {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		w.params.logger.Errorf("%v", err)
	}
	return fmt.Sprintf(w.params.Format, percent[0])
}

func (w *Widget) loop(c chan<- []ygs.I3BarBlock) {
	res := []ygs.I3BarBlock{
		{},
	}
	res[0].FullText = w.getCPU()
	c <- res
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	if w.params.Format == "" {
		w.params.Format = "CPU: %.2f"
	}
	
	w.loop(c)
	ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
	for range ticker.C {
		w.loop(c)
	}

	return nil
}
