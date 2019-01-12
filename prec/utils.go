package prec

func pow10(power int) uint64 {
	r := uint64(1)
	for i := 0; i < power; i++ {
		r *= 10
	}
	return r
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
