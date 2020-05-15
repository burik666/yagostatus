// Package widgets contains a YaGoStatus widgets.
package widgets

import (
	"github.com/burik666/yagostatus/internal/pkg/logger"
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
func NewBlankWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	return &BlankWidget{}, nil
}

// Run starts the main loop.
func (w *BlankWidget) Run(c chan<- []ygs.I3BarBlock) error {
	return nil
}

// Event processes the widget events.
func (w *BlankWidget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) error {
	return nil
}

// Stop stops the widdget.
func (w *BlankWidget) Stop() error {
	return nil
}

// Continue continues the widdget.
func (w *BlankWidget) Continue() error {
	return nil
}

// Shutdown shutdowns the widget.
func (w *BlankWidget) Shutdown() error {
	return nil
}
