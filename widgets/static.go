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

// Configure configures the widget.
func (w *StaticWidget) Configure(cfg map[string]interface{}) error {
	v, ok := cfg["blocks"]
	if !ok {
		return errors.New("missing 'blocks' setting")
	}

	return json.Unmarshal([]byte(v.(string)), &w.blocks)
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
	ygs.RegisterWidget(StaticWidget{})
}
