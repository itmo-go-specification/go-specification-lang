package notation

/*
@
requires acc(a);
requires n == len(a);
requires \forall i integer; (0 <= i) && (i < n) ==> acc(&a[i]) && (a[i] >= 0);
ensures \result >= 0;
*/
func SumPositive(a []int, n int) int {
	s := 0

	/*@ 
	loop invariant acc(a);
    loop invariant (0 <= i) && (i <= n);
    loop invariant \forall j integer; (0 <= j) && (j < n) ==> acc(&a[j]) && a[j] >= 0;
    loop invariant s >= 0 ;
	*/
	for i := 0; i < n; i++ {
		s += a[i]
	}
	return s
}

/*
@
requires acc(a);
requires \forall i integer; (0 <= i) && (i < n) ==> acc(&a[i]);
requires n > 0;
ensures acc(a);
ensures \forall i integer; (0 <= i) && (i < n) ==> acc(&a[i]);
ensures \exists i integer; (0 <= i) && (i < n) && (\result == a[i]);
*/
func FirstElement(a []int, n int) int {
	return a[0]
}
