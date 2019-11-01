package widgets

import (
	"encoding/json"
	"errors"

	"github.com/burik666/yagostatus/ygs"
)

// StaticWidgetParams are widget parameters.
type StaticWidgetParams struct {
	Blocks string
}

// StaticWidget implements a static widget.
type StaticWidget struct {
	params StaticWidgetParams

	blocks []ygs.I3BarBlock
}

func init() {
	ygs.RegisterWidget("static", NewStaticWidget, StaticWidgetParams{})
}

// NewStaticWidget returns a new StaticWidget.
func NewStaticWidget(params interface{}) (ygs.Widget, error) {
	w := &StaticWidget{
		params: params.(StaticWidgetParams),
	}

	if len(w.params.Blocks) == 0 {
		return nil, errors.New("missing 'blocks' setting")
	}

	if err := json.Unmarshal([]byte(w.params.Blocks), &w.blocks); err != nil {
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
func (w *StaticWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) {}

// Stop stops the widdget.
func (w *StaticWidget) Stop() {}

// Continue continues the widdget.
func (w *StaticWidget) Continue() {}

// Shutdown shutdowns the widget.
func (w *StaticWidget) Shutdown() {}
