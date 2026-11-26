package loop

/*
@
requires len(a) > 0;
ensures \result >= a[0];
*/
func max(a []int) int {
	m := a[0]
	i := 1

	/*@
	  loop invariant 1 <= i && i <= len(a);
	  loop assigns i, m;
	  loop variant len(a) - i;
	*/
	for i < len(a) {
		if a[i] > m {
			m = a[i]
		}
		i++
	}
	return m
}
