package ygs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (b *I3BarBlock) FromJSON(data []byte, strict bool) error {
	type dataWrapped I3BarBlock

	var block dataWrapped

	// copy
	tmp, err := json.Marshal(b)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(tmp, &block); err != nil {
		return err
	}

	if block.Custom == nil {
		block.Custom = make(map[string]Vary)
	}

	if err := parseBlock(&block, block.Custom, data, strict); err != nil {
		return err
	}

	*b = I3BarBlock(block)

	return nil
}

func (b *I3BarBlock) ToVaryMap() map[string]Vary {
	tmp, _ := json.Marshal(b)

	varyMap := make(map[string]Vary)

	json.Unmarshal(tmp, &varyMap)

	return varyMap
}

func parseBlock(block interface{}, custom map[string]Vary, data []byte, strict bool) error {
	var jfields map[string]Vary

	if err := json.Unmarshal(data, &jfields); err != nil {
		return err
	}

	val := reflect.ValueOf(block).Elem()
	fieldsByJSONTag := make(map[string]reflect.Value)

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		if tag, ok := typeField.Tag.Lookup("json"); ok {
			tagName := strings.Split(tag, ",")[0]
			if tagName == "-" || tagName == "" {
				continue
			}

			fieldsByJSONTag[tagName] = val.Field(i)
		}
	}

	for k, v := range jfields {
		f, ok := fieldsByJSONTag[k]
		if !ok {
			if len(k) == 0 {
				continue
			}

			if strict && k[0] != byte('_') {
				return fmt.Errorf("uknown field: %s", k)
			}

			custom[k] = v

			continue
		}

		var val reflect.Value

		if f.Type().Kind() == reflect.Ptr {
			val = reflect.New(f.Type().Elem())
		} else {
			val = reflect.New(f.Type()).Elem()
		}

		sv := string(v)

		switch reflect.Indirect(val).Kind() {
		case reflect.String:
			s, err := strconv.Unquote(sv)
			if strict && err != nil {
				return fmt.Errorf("invalid value for %s (string): %s", k, sv)
			} else {
				if err != nil {
					s = sv
				}
			}

			reflect.Indirect(val).SetString(s)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			s := sv
			if !strict {
				s = strings.Trim(sv, "\"")
			}

			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				reflect.Indirect(val).SetInt(n)
			} else {
				return fmt.Errorf("invalid value for %s (int): %s", k, sv)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			s := sv
			if !strict {
				s = strings.Trim(sv, "\"")
			}

			if n, err := strconv.ParseUint(s, 10, 64); err == nil {
				reflect.Indirect(val).SetUint(n)
			} else {
				return fmt.Errorf("invalid value for %s (uint): %s", k, sv)
			}

		case reflect.Bool:
			if strict {
				switch sv {
				case "true":
					reflect.Indirect(val).SetBool(true)
				case "false":
					reflect.Indirect(val).SetBool(false)
				default:
					return fmt.Errorf("invalid value for %s: %s", k, sv)
				}
			} else {
				s := strings.Trim(strings.ToLower(sv), "\"")
				if s == "false" || s == "0" || s == "f" {
					reflect.Indirect(val).SetBool(false)
				} else {
					reflect.Indirect(val).SetBool(true)
				}
			}
		default:
			panic("unsuported type")
		}

		f.Set(val)
	}

	return nil
}
