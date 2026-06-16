package notation

/*
@
requires \forall i integer; (0 <= i) && (i < n) ==> acc(&a[i]) && (a[i] >= 0);
ensures result >= 0;
*/
func SumPositive(a []int, n int) (result int) {
	s := 0
	for i := 0; i < n; i++ {
		s += a[i]
	}
	return s
}

/*
@
requires forall i int :: 0 <= i && i < n ==> acc(&a[i])
ensures \exists i integer; (0 <= i) && (i < n) && (result == a[i]);
*/
func FirstElement(a []int, n int) (result int) {
	return a[0]
}
