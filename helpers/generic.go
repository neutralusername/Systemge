package helpers

func GetNilValue[T any](object T) T {
	var nilValue T
	return nilValue
}

func ChannelCall[T any](f func() (T, error)) <-chan T {
	resultChannel := make(chan T)
	go func() {
		defer close(resultChannel)

		result, err := f()
		if err != nil {
			return
		}
		resultChannel <- result
	}()

	return resultChannel
}
