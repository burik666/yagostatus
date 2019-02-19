// Package ygs contains the YaGoStatus structures.
package ygs

import (
	"bytes"
	"encoding/json"
)

type widgetParams = map[string]interface{}

// Widget represents a widget struct.
type Widget interface {
	Run(chan<- []I3BarBlock) error
	Event(I3BarClickEvent)
	Stop()
}

// UnmarshalJSON unmarshals json with custom keys (with _ prefix).
func (b *I3BarBlock) UnmarshalJSON(data []byte) error {
	type dataWrapped I3BarBlock

	var wr dataWrapped

	if err := json.Unmarshal(data, &wr); err != nil {
		return err
	}

	if err := json.Unmarshal(data, &wr.Custom); err != nil {
		return err
	}
	for k := range wr.Custom {
		if k[0] != '_' {
			delete(wr.Custom, k)
		}
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
