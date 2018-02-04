package widgets

import (
	"encoding/json"
	"errors"
	"github.com/burik666/yagostatus/ygs"
)

type StaticWidget struct {
	blocks []ygs.I3BarBlock
}

func (w *StaticWidget) Configure(cfg map[string]interface{}) error {
	v, ok := cfg["blocks"]
	if !ok {
		return errors.New("Missing 'blocks' setting")
	}

	return json.Unmarshal([]byte(v.(string)), &w.blocks)
}

func (w *StaticWidget) Run(c chan []ygs.I3BarBlock) error {
	c <- w.blocks
	return nil
}

func (w *StaticWidget) Event(event ygs.I3BarClickEvent) {}
func (w *StaticWidget) Stop()                           {}

func init() {
	ygs.RegisterWidget(StaticWidget{})
}
