package tools

func Ternary[T any](condition bool, x, y T) T {
	if condition {
		return x
	}
	return y
}
