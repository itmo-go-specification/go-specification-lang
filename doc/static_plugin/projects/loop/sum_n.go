package loop

/*
@
requires n >= 0;
ensures \result == n*(n+1)/2;
*/
func sumN(n int) int {
	s := 0
	i := 0

	/*@
	  loop invariant 0 <= i && i <= n;
	  loop assigns i, s;
	  loop variant n - i;
	*/
	for i < n {
		i++
		s += i
	}
	return s
}
