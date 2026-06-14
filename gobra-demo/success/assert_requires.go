package success

/*
@
requires x >= 0;
ensures res == x + 1;
assigns \nothing;
*/
func AddOneNegativeSimple(x int) (res int) {
	/*@ assert x >= 0; */
	return x + 1
}
