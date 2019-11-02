package widgets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/burik666/yagostatus/internal/pkg/executor"
	"github.com/burik666/yagostatus/internal/pkg/signals"
	"github.com/burik666/yagostatus/ygs"
)

// ExecWidgetParams are widget parameters.
type ExecWidgetParams struct {
	Command      string
	Interval     int
	Retry        *int
	EventsUpdate bool `yaml:"events_update"`
	Signal       *int
	OutputFormat executor.OutputFormat `yaml:"output_format"`
}

// ExecWidget implements the exec widget.
type ExecWidget struct {
	BlankWidget

	params ExecWidgetParams

	signal       os.Signal
	c            chan<- []ygs.I3BarBlock
	upd          chan struct{}
	customfields map[string]interface{}
	tickerC      *chan struct{}

	outputWG sync.WaitGroup
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
		return nil, errors.New("missing 'command'")
	}

	if w.params.Retry != nil &&
		*w.params.Retry > 0 &&
		w.params.Interval > 0 &&
		*w.params.Retry >= w.params.Interval {
		return nil, errors.New("restart value should be less than interval")
	}

	if w.params.Signal != nil {
		sig := *w.params.Signal
		if sig < 0 || signals.SIGRTMIN+sig > signals.SIGRTMAX {
			return nil, fmt.Errorf("signal should be between 0 AND %d", signals.SIGRTMAX-signals.SIGRTMIN)
		}

		w.signal = syscall.Signal(signals.SIGRTMIN + sig)
	}

	w.upd = make(chan struct{}, 1)
	w.upd <- struct{}{}

	return w, nil
}

func (w *ExecWidget) exec() error {
	exc, err := executor.Exec("sh", "-c", w.params.Command)
	if err != nil {
		return err
	}

	for k, v := range w.customfields {
		vst, _ := json.Marshal(v)
		exc.AddEnv(
			fmt.Sprintf("I3_%s=%s", strings.ToUpper(k), vst),
		)
	}

	c := make(chan []ygs.I3BarBlock)

	defer close(c)

	w.outputWG.Add(1)
	go (func() {
		defer w.outputWG.Done()

		for {
			blocks, ok := <-c
			if !ok {
				return
			}
			w.c <- blocks
			w.setCustomFields(blocks)
		}
	})()

	err = exc.Run(c, w.params.OutputFormat)
	if err == nil {
		if state := exc.ProcessState(); state != nil && state.ExitCode() != 0 {
			if w.params.Retry != nil {
				go (func() {
					time.Sleep(time.Second * time.Duration(*w.params.Retry))
					w.upd <- struct{}{}
					w.resetTicker()
				})()
			}

			return fmt.Errorf("process exited unexpectedly: %s", state.String())
		}
	}

	return err
}

// Run starts the main loop.
func (w *ExecWidget) Run(c chan<- []ygs.I3BarBlock) error {
	w.c = c
	if w.params.Interval == 0 && w.signal == nil && w.params.Retry == nil {
		return w.exec()
	}

	if w.params.Interval > 0 {
		w.resetTicker()
	}

	if w.params.Interval == -1 {
		go (func() {
			for {
				w.upd <- struct{}{}
			}
		})()
	}

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

	for range w.upd {
		err := w.exec()
		if err != nil {
			w.outputWG.Wait()

			c <- []ygs.I3BarBlock{{
				FullText: err.Error(),
				Color:    "#ff0000",
			}}
		}
	}

	return nil
}

// Event processes the widget events.
func (w *ExecWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) error {
	w.setCustomFields(blocks)

	if w.params.EventsUpdate {
		w.upd <- struct{}{}
	}

	return nil
}

func (w *ExecWidget) setCustomFields(blocks []ygs.I3BarBlock) {
	customfields := make(map[string]interface{})

	for _, block := range blocks {
		for k, v := range block.Custom {
			customfields[k] = v
		}
	}

	w.customfields = customfields
}

func (w *ExecWidget) resetTicker() {
	if w.tickerC != nil {
		*w.tickerC <- struct{}{}
	}

	if w.params.Interval > 0 {
		tickerC := make(chan struct{}, 1)
		w.tickerC = &tickerC

		go (func() {
			ticker := time.NewTicker(time.Duration(w.params.Interval) * time.Second)

			defer ticker.Stop()

			for {
				select {
				case <-tickerC:
					return
				case <-ticker.C:
					w.upd <- struct{}{}
				}
			}
		})()
	}
}
