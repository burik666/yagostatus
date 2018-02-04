package ygs

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
	"strings"
)

func (b *I3BarBlock) UnmarshalJSON(data []byte) error {
	type dataWrapped I3BarBlock

	var wr dataWrapped

	if err := json.Unmarshal(data, &wr); err != nil {
		return err
	}

	if err := json.Unmarshal(data, &wr.Custom); err != nil {
		return err
	}
	for k, _ := range wr.Custom {
		if k[0] != '_' {
			delete(wr.Custom, k)
		}
	}

	*b = I3BarBlock(wr)

	return nil
}

func (d I3BarBlock) MarshalJSON() ([]byte, error) {
	type dataWrapped I3BarBlock
	var wd dataWrapped
	wd = dataWrapped(d)

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

	//	return json.Marshal(resmap)
}

var registeredWidgets = make(map[string]reflect.Type)

func RegisterWidget(widget interface{}) error {
	t := reflect.TypeOf(widget)
	name := strings.ToLower(t.Name())
	if _, ok := registeredWidgets[name]; ok {
		log.Fatalf("Widget '%s' already registered", name)
	}
	registeredWidgets[name] = t
	return nil
}

func NewWidget(name string) (Widget, bool) {
	t, ok := registeredWidgets[name]
	if !ok {
		return nil, false
	}
	v := reflect.New(t)
	return v.Interface().(Widget), true
}
