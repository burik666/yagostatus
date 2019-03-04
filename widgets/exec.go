package widgets

import (
	"encoding/json"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/signals"
	"github.com/burik666/yagostatus/ygs"

	"github.com/pkg/errors"
)

// ExecWidget implements the exec widget.
type ExecWidget struct {
	command      string
	interval     time.Duration
	eventsUpdate bool
	signal       os.Signal
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

	v, ok = params["signal"]
	if ok {
		sig := v.(int)
		if sig < 0 || signals.SIGRTMIN+sig > signals.SIGRTMAX {
			return nil, errors.Errorf("signal should be between 0 AND %d", signals.SIGRTMAX-signals.SIGRTMIN)
		}
		w.signal = syscall.Signal(signals.SIGRTMIN + sig)
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
		blocks = append(blocks, ygs.I3BarBlock{FullText: strings.Trim(string(output), "\n ")})
	}
	w.c <- blocks
	return nil

}

// Run starts the main loop.
func (w *ExecWidget) Run(c chan<- []ygs.I3BarBlock) error {
	w.c = c
	if w.interval == 0 && w.signal == nil {
		return w.exec()
	}

	upd := make(chan struct{}, 1)

	if w.signal != nil {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, w.signal)
		go (func() {
			for {
				<-sigc
				upd <- struct{}{}
			}
		})()
	}
	if w.interval > 0 {
		ticker := time.NewTicker(w.interval)
		go (func() {
			for {
				<-ticker.C
				upd <- struct{}{}
			}
		})()
	}

	for ; true; <-upd {
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
