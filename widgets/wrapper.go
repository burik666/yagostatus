package widgets

import (
	"encoding/json"
	"errors"
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
		return nil, errors.New("missing 'command' setting")
	}

	return w, nil
}

// Run starts the main loop.
func (w *WrapperWidget) Run(c chan<- []ygs.I3BarBlock) error {
	var err error
	w.exc, err = executor.Exec(w.params.Command)
	if err != nil {
		return nil
	}
	w.stdin, err = w.exc.Stdin()
	if err != nil {
		return nil
	}
	defer w.stdin.Close()
	w.stdin.Write([]byte("["))
	return w.exc.Run(c, executor.OutputFormatJSON)
}

// Event processes the widget events.
func (w *WrapperWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) {
	if w.stdin != nil {
		msg, _ := json.Marshal(event)
		w.stdin.Write(msg)
		w.stdin.Write([]byte(",\n"))
	}
}

// Stop shutdowns the widget.
func (w *WrapperWidget) Stop() {
	if w.exc != nil {
		w.exc.Signal(syscall.SIGHUP)
		w.exc.Wait()
	}
}
