package main

/*
@
lemma non_negative_successor: \forall i integer; i >= 0 ==> i + 1 > 0;

lemma only_checked_identity: \forall x integer; x == x;

lemma trusted_add_zero: \forall x integer; x + 0 == x;
*/

/*
@
requires n >= 0;
ensures \result >= n;
ensures \result > 0;
*/
func IncNonNegative(n int) int {
	return n + 1
}

/*
@
requires a <= b;
ensures \result >= a;
ensures \result <= b;
*/
func Mid(a int, b int) int {
	return (a + b) / 2
}

func main() {
	_ = IncNonNegative(10)
	_ = Mid(2, 8)
}
