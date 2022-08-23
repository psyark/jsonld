package jsonld

import (
	"encoding/json"
	"fmt"
)

type DecodeOptions struct {
	ExpectText            bool
	ExpectDate            bool
	ExpectDateTime        bool
	ExpectTime            bool
	ExpectURL             bool
	ExpectCssSelectorType bool
	ExpectXPathType       bool
}

func bracket(data []byte) []byte {
	if len(data) == 0 || data[0] != '[' {
		return append(append([]byte{'['}, data...), ']')
	}
	return data
}

func DecodeObject(data []byte, options DecodeOptions) (interface{}, error) {
	if len(data) != 0 && data[0] == '"' {
		var str interface{}
		switch {
		case options.ExpectDate:
			str = &dateStruct{}
		case options.ExpectText:
			str = &textStruct{}
		case options.ExpectURL:
			str = &uRLStruct{}
		default:
			return nil, fmt.Errorf("文字列のクラスを決定できません: %v, %v", options, string(data))
		}

		if err := json.Unmarshal(data, str); err != nil {
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

	var thing Thing = NewThing(typeCheck.Type)
	if err := json.Unmarshal(data, thing); err != nil {
		return nil, err
	}

	return thing, nil
}

func DecodeObjects(data []byte, options DecodeOptions) ([]interface{}, error) {
	items := []json.RawMessage{}
	if err := json.Unmarshal(bracket(data), &items); err != nil {
		return nil, err
	}

	objects := make([]interface{}, len(items))
	for i, msg := range items {
		t, err := DecodeObject(msg, options)
		if err != nil {
			return nil, err
		}
		objects[i] = t
	}

	return objects, nil
}
