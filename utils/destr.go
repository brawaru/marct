package utils

func Unpack[T any](arr []T, vars ...*T) {
	maxLen := len(arr)
	for i, v := range vars {
		if i >= maxLen {
			break
		}

		*v = arr[i]
	}
}
