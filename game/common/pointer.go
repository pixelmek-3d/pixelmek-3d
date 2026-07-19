package common

// Ptr accepts any value, copies it to a local argument, and returns its address
func Ptr[T any](v T) *T {
	return &v
}
