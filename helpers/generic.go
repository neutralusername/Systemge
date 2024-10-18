package helpers

func GetNilValue[T any](object T) T {
	var nilValue T
	return nilValue
}
