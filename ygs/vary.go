package ygs

import (
	"errors"
	"strconv"
)

type Vary []byte

func (v Vary) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return v, nil
}

func (v *Vary) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("ygs.Vary: UnmarshalJSON on nil pointer")
	}
	*v = append((*v)[0:0], data...)
	return nil
}

func (v Vary) String() string {
	s, err := strconv.Unquote(string(v))
	if err != nil {
		return string(v)
	}
	return s
}
