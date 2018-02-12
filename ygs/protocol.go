package ygs

// I3BarHeader represents the header of an i3bar message.
type I3BarHeader struct {
	Version     uint8 `json:"version"`
	StopSignal  int   `json:"stop_signal,omitempty"`
	ContSignal  int   `json:"cont_signal,omitempty"`
	ClickEvents bool  `json:"click_events,omitempty"`
}

// I3BarBlock represents a block of i3bar message.
type I3BarBlock struct {
	FullText            string                 `json:"full_text"`
	ShortText           string                 `json:"short_text,omitempty"`
	Color               string                 `json:"color,omitempty"`
	BorderColor         string                 `json:"border,omitempty"`
	BackgroundColor     string                 `json:"background,omitempty"`
	Markup              string                 `json:"markup,omitempty"`
	MinWidth            string                 `json:"min_width,omitempty"`
	Align               string                 `json:"align,omitempty"`
	Name                string                 `json:"name,omitempty"`
	Instance            string                 `json:"instance,omitempty"`
	Urgent              bool                   `json:"urgent,omitempty"`
	Separator           *bool                  `json:"separator,omitempty"`
	SeparatorBlockWidth uint32                 `json:"separator_block_width,omitempty"`
	Custom              map[string]interface{} `json:"-"`
}

// I3BarClickEvent represents a user click event message.
type I3BarClickEvent struct {
	Name     string `json:"name,omitmpty"`
	Instance string `json:"instance,omitempty"`
	X        uint16 `json:"x"`
	Y        uint16 `json:"y"`
	Button   uint8  `json:"button"`
}
