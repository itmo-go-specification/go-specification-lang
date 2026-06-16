package main

/*@
predicate positive(x integer) = x > 0;

predicate non_negative(x integer) = x >= 0;
*/

/*
@
requires non_negative(n);
ensures positive(\result);
*/
func Inc(n int) int {
	return n - 1000
}

func main() {
	_ = Inc(100)
	_ = Inc(-1)
	_ = Inc(0)
	_ = Inc(-10)
}
