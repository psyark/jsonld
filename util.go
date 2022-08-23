package jsonld

import (
	"encoding/json"
)

type Values []interface{}

func (v Values) Get() interface{} {
	if len(v) != 0 {
		return v[0]
	}
	return nil
}

func (v *Values) UnmarshalJSON(data []byte) error {
	items := []json.RawMessage{}
	if err := json.Unmarshal(bracket(data), &items); err != nil {
		return err
	}

	*v = make([]interface{}, len(items))
	for i, msg := range items {
		t, err := DecodeObject(msg)
		if err != nil {
			return err
		}
		(*v)[i] = t
	}

	return nil
}

func DecodeObject(data []byte) (interface{}, error) {
	var x interface{}
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}

	switch x := x.(type) {
	case map[string]interface{}:
		thing := NewThing(x["@type"].(string))
		if err := json.Unmarshal(data, thing); err != nil {
			return nil, err
		}
		return thing, nil

	default:
		return x, nil
	}
}

func bracket(data []byte) []byte {
	if len(data) == 0 || data[0] != '[' {
		return append(append([]byte{'['}, data...), ']')
	}
	return data
}
