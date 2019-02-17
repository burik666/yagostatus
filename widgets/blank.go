// Package widgets contains a YaGoStatus widgets.
package widgets

import (
	"github.com/burik666/yagostatus/ygs"
)

// BlankWidget is a widgets template.
type BlankWidget struct{}

// Configure configures the widget.
func (w *BlankWidget) Configure(cfg map[string]interface{}) error {
	return nil
}

// Run starts the main loop.
func (w *BlankWidget) Run(c chan<- []ygs.I3BarBlock) error {
	return nil
}

// Event processes the widget events.
func (w *BlankWidget) Event(event ygs.I3BarClickEvent) {}

// Stop shutdowns the widget.
func (w *BlankWidget) Stop() {}

func init() {
	ygs.RegisterWidget(&BlankWidget{})
}
