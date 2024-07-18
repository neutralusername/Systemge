package Helpers

func Max(integers ...int) int {
	numLen := len(integers)
	if numLen == 0 {
		return 0
	}
	max := integers[0]
	for i := 1; i < numLen; i++ {
		if integers[i] > max {
			max = integers[i]
		}
	}
	return max
}

func Min(integers ...int) int {
	numLen := len(integers)
	if numLen == 0 {
		return 0
	}
	min := integers[0]
	for i := 1; i < numLen; i++ {
		if integers[i] < min {
			min = integers[i]
		}
	}
	return min
}

func Absolute(integer int) int {
	if integer < 0 {
		return -integer
	}
	return integer
}
