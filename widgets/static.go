package widgets

import (
	"encoding/json"
	"errors"

	"github.com/burik666/yagostatus/ygs"
)

// StaticWidget implements a static widget.
type StaticWidget struct {
	blocks []ygs.I3BarBlock
}

// NewStaticWidget returns a new StaticWidget.
func NewStaticWidget(params map[string]interface{}) (ygs.Widget, error) {
	w := &StaticWidget{}

	v, ok := params["blocks"]
	if !ok {
		return nil, errors.New("missing 'blocks' setting")
	}

	if err := json.Unmarshal([]byte(v.(string)), &w.blocks); err != nil {
		return nil, err
	}

	return w, nil
}

// Run returns configured blocks.
func (w *StaticWidget) Run(c chan<- []ygs.I3BarBlock) error {
	c <- w.blocks
	return nil
}

// Event processes the widget events.
func (w *StaticWidget) Event(event ygs.I3BarClickEvent) {}

// Stop shutdowns the widget.
func (w *StaticWidget) Stop() {}

func init() {
	ygs.RegisterWidget("static", NewStaticWidget)
}
