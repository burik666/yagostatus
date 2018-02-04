package widgets

import (
	"errors"
	"github.com/burik666/yagostatus/ygs"
	"time"
)

type ClockWidget struct {
	format   string
	interval time.Duration
}

func (w *ClockWidget) Configure(cfg map[string]interface{}) error {
	v, ok := cfg["format"]
	if !ok {
		return errors.New("Missing 'format' setting")
	}
	w.format = v.(string)

	v, ok = cfg["interval"]
	if ok {
		w.interval = time.Duration(v.(int)) * time.Second
	} else {
		w.interval = time.Second
	}

	return nil
}

func (w *ClockWidget) Run(c chan []ygs.I3BarBlock) error {
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
	return nil
}

func (w *ClockWidget) Event(event ygs.I3BarClickEvent) {}
func (w *ClockWidget) Stop()                           {}

func init() {
	ygs.RegisterWidget(ClockWidget{})
}
