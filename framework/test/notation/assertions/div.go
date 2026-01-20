package assertions

func div(x, y int) int {
	/*@ assert y != 0; */
	return x / y
}
