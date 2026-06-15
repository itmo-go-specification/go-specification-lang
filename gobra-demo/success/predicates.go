package success

/*@
predicate positive(x integer) = x > 0;

predicate non_negative(x integer) = x >= 0;
*/

/*
@
requires non_negative(n);
ensures positive(res);
*/
func Inc(n int) (res int) {
	return n + 1
}

func CorrectIncCalls() {
	_ = Inc(100)
	_ = Inc(0)
}
