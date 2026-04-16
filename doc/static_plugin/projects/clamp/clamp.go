package clamp

func clamp(x int, min int, max int) int {
	/*@ assert min <= max; */

	if x < min {
		return min
	}

	/*@ assert x >= min; */

	if x > max {
		return max
	}

	/*@ assert (x >= min) && (x <= max); */
	return x
}
