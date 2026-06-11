package failure

/*
@
requires x >= 0;
ensures res == x + 1;
assigns \nothing;
*/
func AddOneNegativeSimple(x int) (res int) {
	/*@ assert x >= 0; */ // FAILS: x can be negative
	return x + 1
}
