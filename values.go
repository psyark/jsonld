package jsonld

type Values []interface{}

func (v *Values) UnmarshalJSON(data []byte) error {
	return nil
}
