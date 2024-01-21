package pkg

func PointTo[T any](instance T) *T {
	return &instance
}
