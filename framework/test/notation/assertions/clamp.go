package assertions

func clamp(x int) int {
	if x < 0 {
		x = 0
	}
	if x > 100 {
		x = 100
	}

	/*@ assert 0 <= x && x <= 100; */
	return x
}
