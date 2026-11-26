package loop

func isSorted(a []int) int {
	i := 0
	ok := 1

	/*@
	  loop invariant (0 <= i) && (i <= len(a));
	  loop assigns i, ok;
	  loop variant len(a) - i;
	*/
	for i+1 < len(a) {
		if a[i] > a[i+1] {
			ok = 0
			break
		}
		i++
	}
	return ok
}
