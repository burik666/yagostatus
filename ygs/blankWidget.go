package ygs

// BlankWidgetParams are widget parameters.
type BlankWidgetParams struct{}

// BlankWidget is a widgets template.
type BlankWidget struct{}

// NewBlankWidget returns a new BlankWidget.
func NewBlankWidget(params interface{}, wlogger Logger) (Widget, error) {
	return &BlankWidget{}, nil
}

// Run starts the main loop.
func (w *BlankWidget) Run(c chan<- []I3BarBlock) error {
	return nil
}

// Event processes the widget events.
func (w *BlankWidget) Event(event I3BarClickEvent, blocks []I3BarBlock) error {
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
