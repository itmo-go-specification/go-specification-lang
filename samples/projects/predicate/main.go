package main

/*@
predicate positive(x integer) = x > 0;

predicate non_negative(x integer) = x >= 0;

logic inc(x integer) integer = x + 1;

lemma non_negative_inc_positive: \forall x integer; non_negative(x) ==> positive(inc(x));
*/

/*
@
requires non_negative(n);
ensures positive(\result);
*/
func Inc(n int) int {
	//@ assert n > 10
	return n + 1
}

func main() {
	_ = Inc(1)
	_ = Inc(-1)
	_ = Inc(0)
	_ = Inc(-10)
}
