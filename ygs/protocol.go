package ygs

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// I3BarHeader represents the header of an i3bar message.
type I3BarHeader struct {
	Version     uint8 `json:"version"`
	StopSignal  int   `json:"stop_signal,omitempty"`
	ContSignal  int   `json:"cont_signal,omitempty"`
	ClickEvents bool  `json:"click_events,omitempty"`
}

// I3BarBlock represents a block of i3bar message.
type I3BarBlock struct {
	FullText            string          `json:"full_text,omitempty"`
	ShortText           string          `json:"short_text,omitempty"`
	Color               string          `json:"color,omitempty"`
	BorderColor         string          `json:"border,omitempty"`
	BorderTop           *uint16         `json:"border_top,omitempty"`
	BorderBottom        *uint16         `json:"border_bottom,omitempty"`
	BorderLeft          *uint16         `json:"border_left,omitempty"`
	BorderRight         *uint16         `json:"border_right,omitempty"`
	BackgroundColor     string          `json:"background,omitempty"`
	Markup              string          `json:"markup,omitempty"`
	MinWidth            string          `json:"min_width,omitempty"`
	Align               string          `json:"align,omitempty"`
	Name                string          `json:"name,omitempty"`
	Instance            string          `json:"instance,omitempty"`
	Urgent              bool            `json:"urgent,omitempty"`
	Separator           *bool           `json:"separator,omitempty"`
	SeparatorBlockWidth uint16          `json:"separator_block_width,omitempty"`
	Custom              map[string]Vary `json:"-"`
}

// I3BarClickEvent represents a user click event message.
type I3BarClickEvent struct {
	Name      string   `json:"name,omitempty"`
	Instance  string   `json:"instance,omitempty"`
	Button    uint8    `json:"button"`
	X         uint16   `json:"x"`
	Y         uint16   `json:"y"`
	RelativeX uint16   `json:"relative_x"`
	RelativeY uint16   `json:"relative_y"`
	Width     uint16   `json:"width"`
	Height    uint16   `json:"height"`
	Modifiers []string `json:"modifiers"`
}

// UnmarshalJSON unmarshals json with custom keys (with _ prefix).
func (b *I3BarBlock) UnmarshalJSON(data []byte) error {
	return b.FromJSON(data, true)
}

// MarshalJSON marshals json with custom keys (with _ prefix).
func (b I3BarBlock) MarshalJSON() ([]byte, error) {
	type dataWrapped I3BarBlock

	wd := dataWrapped(b)

	if len(wd.Custom) == 0 {
		buf := &bytes.Buffer{}
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(wd)

		return buf.Bytes(), err
	}

	var resmap map[string]interface{}

	var tmp []byte

	tmp, _ = json.Marshal(wd)
	if err := json.Unmarshal(tmp, &resmap); err != nil {
		return nil, err
	}

	tmp, _ = json.Marshal(wd.Custom)
	if err := json.Unmarshal(tmp, &resmap); err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(resmap)

	return buf.Bytes(), err
}

func (b *I3BarBlock) Apply(tpl I3BarBlock) {
	jb, _ := json.Marshal(b)
	*b = tpl

	json.Unmarshal(jb, b)
}

func (b I3BarBlock) Env(suffix string) []string {
	env := make([]string, 0)
	for k, v := range b.Custom {
		env = append(env, fmt.Sprintf("I3_%s%s=%s", k, suffix, v))
	}

	ob, _ := json.Marshal(b)

	var rawOutput map[string]Vary
	json.Unmarshal(ob, &rawOutput)

	for k, v := range rawOutput {
		env = append(env, fmt.Sprintf("I3_%s%s=%s", k, suffix, v.String()))
	}

	return env
}
