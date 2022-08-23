package codegen

import (
	"encoding/json"
)

type Document struct {
	Graph []Graph `json:"@graph"`
}

type Graph struct {
	ID             string          `json:"@id"`
	Type           Strings         `json:"@type"`
	Comment        string          `json:"rdfs:comment"`
	SubClassOf     Refs            `json:"rdfs:subClassOf"`
	SubPropertyOf  json.RawMessage `json:"rdfs:subPropertyOf"`
	DomainIncludes Refs            `json:"schema:domainIncludes"`
	InverseOf      json.RawMessage `json:"schema:inverseOf"`
	IsPartOf       json.RawMessage `json:"schema:isPartOf"`
	RangeIncludes  Refs            `json:"schema:rangeIncludes"`
	SameAs         json.RawMessage `json:"schema:sameAs"`
	Source         json.RawMessage `json:"schema:source"`
	SupersededBy   json.RawMessage `json:"schema:supersededBy"`
	// Label          string          `json:"rdfs:label"`
}

func (g Graph) HasType(typeName string) bool {
	for _, t := range g.Type {
		if t == typeName {
			return true
		}
	}
	return false
}

type Ref struct {
	ID string `json:"@id"`
}

type Strings []string

func (s *Strings) UnmarshalJSON(data []byte) error {
	type alias Strings
	return json.Unmarshal(bracket(data), (*alias)(s))
}

type Refs []Ref

func (r *Refs) UnmarshalJSON(data []byte) error {
	type alias Refs
	return json.Unmarshal(bracket(data), (*alias)(r))
}
