package clock

import (
	"github.com/burik666/yagostatus/widgets/blank"
	"strings"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
	"github.com/mdlayher/wifi"
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
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("wifi", NewWidget, WidgetParams{
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

func (w *Widget) loop(c chan<- []ygs.I3BarBlock) {
	res := []ygs.I3BarBlock{
		{},
	}
	res[0].FullText = w.getWifi()
	c <- res
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	if w.params.Format == "" {
		w.params.Format = "%s"
	}

	w.loop(c)
	ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
	for range ticker.C {
		w.loop(c)
	}

	return nil
}

func (w *Widget) getWifi() string {
	if w.params.iface == nil && w.params.c == nil {
		client, err := wifi.New()
		if err != nil {
			w.params.logger.Errorf("unable to start wifi client: %v", err)
			return ""
		}
		w.params.c = client
	}
	
	if w.params.iface == nil {
		interfaces, err := w.params.c.Interfaces()
		if err != nil {
			w.params.logger.Errorf("unable to get wifi interfaces: %v", err)
			return ""
		}

		for _, i := range interfaces {
			w.params.logger.Infof("interface: %v", i.Name)
			if i.Name == w.params.Interface {
				w.params.iface = i
				break
			}
		}
		
		if w.params.iface == nil {
			w.params.logger.Errorf("wifi interface not found")
			return ""
		}
	}

	bss, err := w.params.c.BSS(w.params.iface)
	if err != nil {
		return "Not connected"
	}
	
	return formatBSS(w.params.Format, bss)

}

func formatBSS(format string, bss *wifi.BSS) string {
	format = strings.Replace(format, "%b", bss.BSSID.String(), -1)
	format = strings.Replace(format, "%s", bss.SSID, -1)
	return format
}
