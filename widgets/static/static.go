package static

import (
	"encoding/json"
	"errors"
	"github.com/burik666/yagostatus/widgets/blank"
	"github.com/burik666/yagostatus/internal/pkg/logger"
	"github.com/burik666/yagostatus/ygs"
)

// StaticWidgetParams are widget parameters.
type StaticWidgetParams struct {
	Blocks string
}

// StaticWidget implements a static widget.
type StaticWidget struct {
	blank.Widget
	params StaticWidgetParams

	blocks []ygs.I3BarBlock
}

var _ ygs.Widget = &StaticWidget{}

func init() {
	ygs.RegisterWidget("static", NewStaticWidget, StaticWidgetParams{})
}

// NewStaticWidget returns a new StaticWidget.
func NewStaticWidget(params interface{}, wlogger logger.Logger) (ygs.Widget, error) {
	w := &StaticWidget{
		params: params.(StaticWidgetParams),
	}

	if len(w.params.Blocks) == 0 {
		return nil, errors.New("missing 'blocks'")
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
