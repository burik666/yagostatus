package ygs

import (
	"bytes"
	"encoding/json"
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
	FullText            string                 `json:"full_text"`
	ShortText           string                 `json:"short_text,omitempty"`
	Color               string                 `json:"color,omitempty"`
	BorderColor         string                 `json:"border,omitempty"`
	BorderTop           *uint16                `json:"border_top,omitempty"`
	BorderBottom        *uint16                `json:"border_bottom,omitempty"`
	BorderLeft          *uint16                `json:"border_left,omitempty"`
	BorderRight         *uint16                `json:"border_right,omitempty"`
	BackgroundColor     string                 `json:"background,omitempty"`
	Markup              string                 `json:"markup,omitempty"`
	MinWidth            string                 `json:"min_width,omitempty"`
	Align               string                 `json:"align,omitempty"`
	Name                string                 `json:"name,omitempty"`
	Instance            string                 `json:"instance,omitempty"`
	Urgent              bool                   `json:"urgent,omitempty"`
	Separator           *bool                  `json:"separator,omitempty"`
	SeparatorBlockWidth uint16                 `json:"separator_block_width,omitempty"`
	Custom              map[string]interface{} `json:"-"`
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
	var resmap map[string]interface{}

	if err := json.Unmarshal(data, &resmap); err != nil {
		return err
	}

	type dataWrapped I3BarBlock
	wr := dataWrapped(*b)

	for k, v := range resmap {
		if len(k) == 0 {
			delete(resmap, k)
			continue
		}

		if k[0] == '_' {
			if wr.Custom == nil {
				wr.Custom = make(map[string]interface{})
			}
			wr.Custom[k] = v
			delete(resmap, k)
		}
	}

	buf := &bytes.Buffer{}

	enc := json.NewEncoder(buf)
	enc.Encode(resmap)

	decoder := json.NewDecoder(buf)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wr); err != nil {
		return err
	}

	*b = I3BarBlock(wr)

	return nil
}

// MarshalJSON marshals json with custom keys (with _ prefix).
func (b I3BarBlock) MarshalJSON() ([]byte, error) {
	type dataWrapped I3BarBlock
	var wd dataWrapped
	wd = dataWrapped(b)

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
	json.Unmarshal(tmp, &resmap)

	tmp, _ = json.Marshal(wd.Custom)
	json.Unmarshal(tmp, &resmap)

	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(resmap)
	return buf.Bytes(), err
}
