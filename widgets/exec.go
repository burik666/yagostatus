package widgets

import (
	"encoding/json"
	"errors"
	"github.com/burik666/yagostatus/ygs"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ExecWidget struct {
	command       string
	interval      time.Duration
	events_update bool
	c             chan []ygs.I3BarBlock
}

func (w *ExecWidget) Configure(cfg map[string]interface{}) error {
	v, ok := cfg["command"]
	if !ok {
		return errors.New("Missing 'command' setting")
	}
	w.command = v.(string)

	v, ok = cfg["interval"]
	if !ok {
		return errors.New("Missing 'interval' setting")
	}
	w.interval = time.Second * time.Duration(v.(int))

	v, ok = cfg["events_update"]
	if ok {
		w.events_update = v.(bool)
	} else {
		w.events_update = false
	}

	return nil
}

func (w *ExecWidget) exec() error {
	cmd := exec.Command("sh", "-c", w.command)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var blocks []ygs.I3BarBlock
	err = json.Unmarshal(output, &blocks)
	if err != nil {
		blocks = append(blocks, ygs.I3BarBlock{FullText: strings.Trim(string(output), "\n ")})
	}
	w.c <- blocks
	return nil

}

func (w *ExecWidget) Run(c chan []ygs.I3BarBlock) error {
	w.c = c
	if w.interval == 0 {
		return w.exec()
	}

	ticker := time.NewTicker(w.interval)

	for ; true; <-ticker.C {
		err := w.exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *ExecWidget) Event(event ygs.I3BarClickEvent) {
	if w.events_update {
		w.exec()
	}
}

func (w *ExecWidget) Stop() {}

func init() {
	ygs.RegisterWidget(ExecWidget{})
}
