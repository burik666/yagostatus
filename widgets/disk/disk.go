package disk

import (
	"fmt"
	"github.com/burik666/yagostatus/widgets/blank"
	"syscall"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
)

type WidgetParams struct {
	Interval uint
	logger   logger.Logger
	Format   string
	Fs       string
}

type Widget struct {
	blank.Widget
	params WidgetParams
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("disk", NewWidget, WidgetParams{
		Interval: 1,
	})
}

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
	res[0].FullText = w.getDisk()
	c <- res
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	if w.params.Format == "" {
		w.params.Format = "Disk: %.2f"
	}

	w.loop(c)
	ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
	for range ticker.C {
		w.loop(c)
	}

	return nil
}

// From: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
// Thanks to Stefan Nilsson
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func (w *Widget) getDisk() string {
	var statFS syscall.Statfs_t
	err := syscall.Statfs(w.params.Fs, &statFS)

	if err != nil {
		w.params.logger.Errorf("unable to get fs stats: %v", err)
		return "N/A"
	}

	return fmt.Sprintf(w.params.Format, ByteCountIEC(int64(statFS.Bavail) * statFS.Bsize))
}
