package codegen

func bracket(data []byte) []byte {
	if len(data) == 0 || data[0] != '[' {
		return append(append([]byte{'['}, data...), ']')
	}
	return data
}
