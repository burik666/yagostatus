// Package widgets contains a YaGoStatus widgets.
package blank

import (
	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
)

// BlankWidgetParams are widget parameters.
type WidgetParams struct{}

// BlankWidget is a widgets template.
type Widget struct{}

var _ ygs.Widget = &Widget{}

func init() {
	ygs.RegisterWidget("blank", NewBlankWidget, WidgetParams{})
}

// NewBlankWidget returns a new BlankWidget.
func NewBlankWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	return &Widget{}, nil
}

// Run starts the main loop.
func (w *Widget) Run(c chan<- []ygs.I3BarBlock) error {
	return nil
}

// Event processes the widget events.
func (w *Widget) Event(event ygs.I3BarClickEvent, blocks []ygs.I3BarBlock) error {
	return nil
}

// Stop stops the widget.
func (w *Widget) Stop() error {
	return nil
}

// Continue continues the widget.
func (w *Widget) Continue() error {
	return nil
}

// Shutdown shutdowns the widget.
func (w *Widget) Shutdown() error {
	return nil
}
