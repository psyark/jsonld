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
	return decodeObjectsAs((*[]interface{})(v), data, "")
}

func decodeObjectsAs(dst *[]interface{}, data []byte, defType string) error {
	items := []json.RawMessage{}
	if err := json.Unmarshal(bracket(data), &items); err != nil {
		return err
	}

	*dst = make([]interface{}, len(items))
	for i, msg := range items {
		t, err := decodeObjectAs(msg, defType)
		if err != nil {
			return err
		}
		(*dst)[i] = t
	}

	return nil
}

func decodeObjectAs(data []byte, defType string) (interface{}, error) {
	var x interface{}
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}

	switch x := x.(type) {
	case map[string]interface{}:
		typeName := defType
		if tn, ok := x["@type"].(string); ok {
			typeName = tn
		}

		thing, err := NewThing(typeName)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, thing); err != nil {
			return nil, err
		}
		return thing, nil

	default:
		return x, nil
	}
}

func DecodeObject(data []byte) (interface{}, error) {
	return decodeObjectAs(data, "")
}

func bracket(data []byte) []byte {
	if len(data) == 0 || data[0] != '[' {
		return append(append([]byte{'['}, data...), ']')
	}
	return data
}
