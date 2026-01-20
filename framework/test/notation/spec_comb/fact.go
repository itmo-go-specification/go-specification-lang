package spec_comb

/*
@
requires n >= 0;
ensures \result >= 1;
*/
func fact(n int) int {
	res := 1
	i := 1

	//@ assert res == 1;

	/*@
	  loop invariant 1 <= i && i <= n+1;
	  loop invariant res >= 1;
	  loop assigns i, res;
	  loop variant (n+1) - i;
	*/
	for i <= n {
		//@ assert res >= 1;
		res = res * i
		//@ assert res >= 1;
		i++
	}

	//@ assert i == n+1;
	return res
}
