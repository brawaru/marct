package strutils

// Ref returns reference to a string.
func Ref(v string) *string {
	return &v
}
