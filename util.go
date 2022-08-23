package jsonld

import (
	"encoding/json"
)

func bracket(data []byte) []byte {
	if len(data) == 0 || data[0] != '[' {
		return append(append([]byte{'['}, data...), ']')
	}
	return data
}

func DecodeObject(data []byte) (interface{}, error) {
	if len(data) != 0 && data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return nil, err
		}
		return str, nil
	}

	typeCheck := struct {
		Type string `json:"@type"`
	}{}
	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return nil, err
	}

	thing := NewThing(typeCheck.Type)
	if err := json.Unmarshal(data, thing); err != nil {
		return nil, err
	}

	return thing, nil
}

func DecodeObjects(data []byte) ([]interface{}, error) {
	items := []json.RawMessage{}
	if err := json.Unmarshal(bracket(data), &items); err != nil {
		return nil, err
	}

	objects := make([]interface{}, len(items))
	for i, msg := range items {
		t, err := DecodeObject(msg)
		if err != nil {
			return nil, err
		}
		objects[i] = t
	}

	return objects, nil
}
