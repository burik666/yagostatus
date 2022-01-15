package mobilebroadband

import (
	"fmt"
	mb "github.com/denysvitali/go-mobilebroadband"
	"github.com/denysvitali/yagostatus/widgets/blank"
	"time"

	"github.com/denysvitali/yagostatus/internal/pkg/logger"
	"github.com/denysvitali/yagostatus/ygs"
)

type WidgetParams struct {
	Interval  uint
	Format    string
	Interface string
	logger    logger.Logger
}

type Widget struct {
	blank.Widget
	params          WidgetParams
	mobileBroadband *mb.MobileBroadband
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("mobilebroadband", NewWidget, WidgetParams{
		Interval: 1,
	})
}

// NewWidget returns a new MobileBroadband Widet.
func NewWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	m, err := mb.New()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize MobileBroadband library: %v", err)
	}
	w := &Widget{
		params:          params.(WidgetParams),
		mobileBroadband: m,
	}
	w.params.logger = wlogger

	return w, nil
}

func (w *Widget) loop(c chan<- []ygs.I3BarBlock) {
	res := []ygs.I3BarBlock{
		{},
	}
	res[0].FullText = fmt.Sprintf(w.params.Format, w.status())
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

func getTechnology(tech int) string {
	switch mb.AccessTechnology(tech) {
	case mb.EVDOAAt, mb.EVDOBAt, mb.EVDO0At:
		return "E"
	case mb.GPRSAt:
		return "G"
	case mb.GSMAt, mb.GSMCompactAt:
		return "GSM"
	case mb.HSPAAt, mb.HSDPAAt, mb.HSUPAAt:
		return "H+"
	case mb.OneXRTTAt:
		return "3G"
	case mb.LTEAt:
		return "4G"
	case mb.FiveGNRAt:
		return "5G"
	}
	return "?"
}

func (w *Widget) status() string {
	// Select first modem
	modems, err := w.mobileBroadband.Modems()
	if err != nil {
		w.params.logger.Errorf("unable to get modems: %v", err)
		return "ERR"
	}
	if len(modems) == 0 {
		return "N/A"
	}

	first := modems[0]
	status, err := first.SimpleStatus()
	if err != nil {
		w.params.logger.Errorf("unable to get simplestatus: %v", err)
		return "ERR"
	}

	return fmt.Sprintf("%s (%s) - %.0f%%",
		status.OperatorName,
		getTechnology(status.AccessTechnologies),
		status.SignalQuality.Value,
	)

}
