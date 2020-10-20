package clock

import (
	"fmt"
	"github.com/burik666/yagostatus/widgets/blank"
	"strings"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
	"github.com/distatus/battery"
)

type WidgetParams struct {
	Interval uint
	logger   logger.Logger
	Index    int
	Format   string
}

type Widget struct {
	blank.Widget
	params WidgetParams
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("battery", NewWidget, WidgetParams{
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

func (w *Widget) getBattery() string {
	batteries, err := battery.GetAll()
	if err != nil {
		w.params.logger.Errorf("unable to get batteries: %v", err)
		return "N/A"
	}

	if w.params.Index >= len(batteries) || w.params.Index < 0 {
		w.params.logger.Errorf(
			"invalid battery index provided, detected %d batteries",
			len(batteries),
		)
		return "N/A"
	}
	
	if w.params.Format == "" {
		w.params.Format = "%e"
	}
	
	
	batt := batteries[w.params.Index]
	
	return formatBattery(w.params.Format, batt)
}

func formatBattery(format string, batt *battery.Battery) string {
	format = strings.Replace(format, "%e", getStateEmoji(batt), -1)
	format = strings.Replace(format, "%p", getPercentage(batt), -1)
	format = strings.Replace(format, "%v", getVoltage(batt), -1)
	return format
}

func getVoltage(batt *battery.Battery) string {
	return fmt.Sprintf("%.2f", batt.Voltage)
}


func getPercentage(batt *battery.Battery) string {
	return fmt.Sprintf("%.2f", batt.Current/batt.Full * 100)
}

func getStateEmoji(batt *battery.Battery) string {
	switch batt.State {
	case battery.Charging:
		return "🔺"
	case battery.Discharging:
		return "🔻"
	case battery.Full:
		return "🔋"
	case battery.Empty:
		return "😥"
	case battery.NotCharging:
		return "❌"
	case battery.Unknown:
		return "❓"
	}

	return "❓"
}

func (w *Widget) loop(c chan<- []ygs.I3BarBlock) {
	res := []ygs.I3BarBlock{
		{},
	}
	res[0].FullText = w.getBattery()
	c <- res
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	w.loop(c)
	ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
	for range ticker.C {
		w.loop(c)
	}

	return nil
}
