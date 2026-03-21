package assertions

func normalize(x int) int {
	if x < 0 {
		x = -x
	}

	/*@ assert x >= 0;*/
	return x
}
