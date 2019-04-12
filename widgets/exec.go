package widgets

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/executor"
	"github.com/burik666/yagostatus/internal/pkg/signals"
	"github.com/burik666/yagostatus/ygs"
)

// ExecWidgetParams are widget parameters.
type ExecWidgetParams struct {
	Command      string
	Interval     uint
	EventsUpdate bool `yaml:"events_update"`
	Signal       *int
	OutputFormat executor.OutputFormat `yaml:"output_format"`
}

// ExecWidget implements the exec widget.
type ExecWidget struct {
	params ExecWidgetParams

	signal os.Signal
	c      chan<- []ygs.I3BarBlock
	upd    chan struct{}
}

func init() {
	ygs.RegisterWidget("exec", NewExecWidget, ExecWidgetParams{})
}

// NewExecWidget returns a new ExecWidget.
func NewExecWidget(params interface{}) (ygs.Widget, error) {
	w := &ExecWidget{
		params: params.(ExecWidgetParams),
	}

	if len(w.params.Command) == 0 {
		return nil, errors.New("missing 'command' setting")
	}

	if w.params.Signal != nil {
		sig := *w.params.Signal
		if sig < 0 || signals.SIGRTMIN+sig > signals.SIGRTMAX {
			return nil, fmt.Errorf("signal should be between 0 AND %d", signals.SIGRTMAX-signals.SIGRTMIN)
		}
		w.signal = syscall.Signal(signals.SIGRTMIN + sig)
	}
	return w, nil
}

func (w *ExecWidget) exec() error {
	exc, err := executor.Exec("sh", "-c", w.params.Command)
	if err != nil {
		return err
	}
	return exc.Run(w.c, w.params.OutputFormat)
}

// Run starts the main loop.
func (w *ExecWidget) Run(c chan<- []ygs.I3BarBlock) error {
	w.c = c
	if w.params.Interval == 0 && w.signal == nil {
		return w.exec()
	}

	w.upd = make(chan struct{}, 1)

	if w.signal != nil {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, w.signal)
		go (func() {
			for {
				<-sigc
				w.upd <- struct{}{}
			}
		})()
	}
	if w.params.Interval > 0 {
		ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)
		go (func() {
			for {
				<-ticker.C
				w.upd <- struct{}{}
			}
		})()
	}

	for ; true; <-w.upd {
		err := w.exec()
		if err != nil {
			w.c <- []ygs.I3BarBlock{
				ygs.I3BarBlock{
					FullText: err.Error(),
					Color:    "#ff0000",
				},
			}
		}
	}
	return nil
}

// Event processes the widget events.
func (w *ExecWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) {
	if w.params.EventsUpdate {
		w.upd <- struct{}{}
	}
}

// Stop shutdowns the widget.
func (w *ExecWidget) Stop() {}
