package codegen

import (
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed schemaorg-current-https.jsonld
var jsonBytes []byte

func TestBuilder(t *testing.T) {
	d := Document{}
	json.Unmarshal(jsonBytes, &d)

	if err := NewBuilder(d).Build(); err != nil {
		t.Fatal(err)
	}
}
