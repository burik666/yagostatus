// Package widgets contains a YaGoStatus widgets.
package widgets

import (
	"github.com/burik666/yagostatus/ygs"
)

// BlankWidgetParams are widget parameters.
type BlankWidgetParams struct{}

// BlankWidget is a widgets template.
type BlankWidget struct{}

func init() {
	ygs.RegisterWidget("blank", NewBlankWidget, BlankWidgetParams{})
}

// NewBlankWidget returns a new BlankWidget.
func NewBlankWidget(params interface{}) (ygs.Widget, error) {
	return &BlankWidget{}, nil
}

// Run starts the main loop.
func (w *BlankWidget) Run(c chan<- []ygs.I3BarBlock) error {
	return nil
}

// Event processes the widget events.
func (w *BlankWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) {}

// Stop shutdowns the widget.
func (w *BlankWidget) Stop() {}
