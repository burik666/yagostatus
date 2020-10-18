package brightness

import (
	"fmt"
	"github.com/burik666/yagostatus/widgets/blank"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
)

const NotAvailable = "N/A"

type WidgetParams struct {
	Interval uint
	logger   logger.Logger
	Format   string
	Device   string
}

type Widget struct {
	blank.Widget
	params WidgetParams
}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("brightness", NewWidget, WidgetParams{
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
	res[0].FullText = w.getBrightness()
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

func readDevice(devName string, property string) (int, error) {
	f, err := os.Open(fmt.Sprintf("/sys/class/backlight/%s/%s", devName, property))
	if err != nil {
		return -1, err
	}
	
	
	v, err := ioutil.ReadAll(f)
	if err != nil {
		return -1, err
	}
	
	strVal := strings.Replace(string(v), "\n", "", -1)
	
	value, err := strconv.Atoi(strVal)
	if err != nil {
		return -1, err
	}
	
	return value, nil
	
}

func (w *Widget) getBrightness() string {
	actual, err := readDevice(w.params.Device, "brightness")
	if err != nil {
		w.params.logger.Errorf("unable to get brightness: %v", err)
		return NotAvailable
	}
	
	max, err := readDevice(w.params.Device, "max_brightness")
	if err != nil {
		w.params.logger.Errorf("unable to get max brightness: %v", err)
		return NotAvailable
	}


	return formatBrightness(w.params.Format, float64(actual)/float64(max) * 100)

}

func formatBrightness(format string, brightness float64) string {
	return fmt.Sprintf(format, brightness)
}
