package widgets

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/burik666/yagostatus/ygs"
)

// ExecWidget implements the exec widget.
type ExecWidget struct {
	command      string
	interval     time.Duration
	eventsUpdate bool
	c            chan<- []ygs.I3BarBlock
}

// NewExecWidget returns a new ExecWidget.
func NewExecWidget(params map[string]interface{}) (ygs.Widget, error) {
	w := &ExecWidget{}

	v, ok := params["command"]
	if !ok {
		return nil, errors.New("missing 'command' setting")
	}
	w.command = v.(string)

	v, ok = params["interval"]
	if !ok {
		return nil, errors.New("missing 'interval' setting")
	}
	w.interval = time.Second * time.Duration(v.(int))

	v, ok = params["events_update"]
	if ok {
		w.eventsUpdate = v.(bool)
	} else {
		w.eventsUpdate = false
	}

	return w, nil
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
		log.Printf("Failed to parse output: %s", err)
		blocks = append(blocks, ygs.I3BarBlock{FullText: strings.Trim(string(output), "\n ")})
	}
	w.c <- blocks
	return nil

}

// Run starts the main loop.
func (w *ExecWidget) Run(c chan<- []ygs.I3BarBlock) error {
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

// Event processes the widget events.
func (w *ExecWidget) Event(event ygs.I3BarClickEvent) {
	if w.eventsUpdate {
		w.exec()
	}
}

// Stop shutdowns the widget.
func (w *ExecWidget) Stop() {}

func init() {
	ygs.RegisterWidget("exec", NewExecWidget)
}
