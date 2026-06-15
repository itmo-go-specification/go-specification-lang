package failure

/*@
predicate positive(x integer) = x > 0;

predicate non_negative(x integer) = x >= 0;
*/

/*
@
requires non_negative(n);
ensures positive(res);
*/
func WrongInc(n int) (res int) {
	return n - 1000
}

func CallWrongInc() {
	_ = WrongInc(100)
	_ = WrongInc(-1)
	_ = WrongInc(0)
	_ = WrongInc(-10)
}
