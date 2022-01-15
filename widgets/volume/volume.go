package wifi

import (
	"fmt"
	"github.com/denysvitali/yagostatus/internal/pkg/logger"
	"github.com/denysvitali/yagostatus/widgets/blank"
	"github.com/denysvitali/yagostatus/ygs"
	"github.com/mdlayher/wifi"
	"mrogalski.eu/go/pulseaudio"
	"strings"
)

type WidgetParams struct {
	Interval  uint
	logger    logger.Logger
	Format    string
	Interface string
	c         *wifi.Client
	iface     *wifi.Interface
}

type Widget struct {
	blank.Widget
	params WidgetParams
	client *pulseaudio.Client
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("volume", NewWidget, WidgetParams{
		Interval: 1,
	})
}

// NewWidget returns a new ClockWidget.
func NewWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return nil, fmt.Errorf("unable to get pulseaudio client")
	}

	w := &Widget{
		params: params.(WidgetParams),
		client: client,
	}
	w.params.logger = wlogger

	return w, nil
}

func (w *Widget) loop(c chan<- []ygs.I3BarBlock) {
	res := []ygs.I3BarBlock{
		{},
	}
	res[0].FullText = w.getVolume()
	c <- res
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	if w.params.Format == "" {
		w.params.Format = "%l"
	}

	updates, err := w.client.Updates()
	if err != nil {
		return err
	}

	w.loop(c)
	for range updates {
		w.loop(c)
	}

	return nil
}

func (w *Widget) getVolume() string {
	v, err := w.client.Volume()
	if err != nil {
		return "ERR"
	}
	return formatOutput(w.params.Format, v)

}

func formatOutput(format string, volume float32) string {
	format = strings.Replace(format, "%l", fmt.Sprintf("%.2f", volume*100), -1)
	return format
}
