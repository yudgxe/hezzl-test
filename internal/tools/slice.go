package tools

func MergeSlices[T any](base []T, add []T, index int) []T {
	if len(base) == 0 {
		return add
	}

	if index < 0 {
		index = 0
	}

	response := make([]T, 0, len(base)+len(add))
	if len(base) <= index {
		response = append(response, base...)
		response = append(response, add...)
	} else {
		response = append(response, base[:index]...)
		response = append(response, add...)
		response = append(response, base[index:]...)
	}
	return response
}
