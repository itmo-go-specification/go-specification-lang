package notation

/*
@
requires \forall i integer; (0 <= i) && (i < n) ==> (a[i] >= 0);
ensures \result >= 0;
*/
func SumPositive(a []int, n int) int {
	s := 0
	for i := 0; i < n; i++ {
		s += a[i]
	}
	return s
}

/*
@
ensures \exists i integer; (0 <= i) && (i < n) && (\result == a[i]);
*/
func FirstElement(a []int, n int) int {
	return a[0]
}
