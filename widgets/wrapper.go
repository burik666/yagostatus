package widgets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"syscall"

	"github.com/burik666/yagostatus/internal/pkg/executor"
	"github.com/burik666/yagostatus/ygs"
)

// WrapperWidgetParams are widget parameters.
type WrapperWidgetParams struct {
	Command string
}

// WrapperWidget implements the wrapper of other status commands.
type WrapperWidget struct {
	params WrapperWidgetParams

	exc   *executor.Executor
	stdin io.WriteCloser

	eventBracketWritten bool
}

func init() {
	ygs.RegisterWidget("wrapper", NewWrapperWidget, WrapperWidgetParams{})
}

// NewWrapperWidget returns a new WrapperWidget.
func NewWrapperWidget(params interface{}) (ygs.Widget, error) {
	w := &WrapperWidget{
		params: params.(WrapperWidgetParams),
	}

	if len(w.params.Command) == 0 {
		return nil, errors.New("missing 'command'")
	}

	exc, err := executor.Exec(w.params.Command)
	if err != nil {
		return nil, err
	}

	w.exc = exc

	return w, nil
}

// Run starts the main loop.
func (w *WrapperWidget) Run(c chan<- []ygs.I3BarBlock) error {
	var err error

	w.stdin, err = w.exc.Stdin()
	if err != nil {
		return err
	}

	defer w.stdin.Close()

	err = w.exc.Run(c, executor.OutputFormatJSON)
	if err == nil {
		err = errors.New("process exited unexpectedly")

		if state := w.exc.ProcessState(); state != nil {
			return fmt.Errorf("%w: %s", err, state.String())
		}
	}

	return err
}

// Event processes the widget events.
func (w *WrapperWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) error {
	if w.stdin == nil {
		return nil
	}

	if header := w.exc.I3BarHeader(); header != nil && header.ClickEvents {
		if !w.eventBracketWritten {
			w.eventBracketWritten = true
			if _, err := w.stdin.Write([]byte("[")); err != nil {
				return err
			}
		}

		msg, err := json.Marshal(event)
		if err != nil {
			return err
		}

		msg = append(msg, []byte(",\n")...)

		if _, err := w.stdin.Write(msg); err != nil {
			return err
		}
	}

	return nil
}

// Stop stops the widdget.
func (w *WrapperWidget) Stop() error {
	if header := w.exc.I3BarHeader(); header != nil {
		if header.StopSignal != 0 {
			return w.exc.Signal(syscall.Signal(header.StopSignal))
		}
	}

	return w.exc.Signal(syscall.SIGSTOP)
}

// Continue continues the widdget.
func (w *WrapperWidget) Continue() error {
	if header := w.exc.I3BarHeader(); header != nil {
		if header.ContSignal != 0 {
			return w.exc.Signal(syscall.Signal(header.ContSignal))
		}
	}

	return w.exc.Signal(syscall.SIGCONT)
}

// Shutdown shutdowns the widget.
func (w *WrapperWidget) Shutdown() error {
	if w.exc != nil {
		err := w.exc.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}

		return w.exc.Wait()
	}

	return nil
}
