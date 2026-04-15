//go:build exclude

package contracts

/*
@
assigns x;
ensures x == \old(x) + 1;
*/
func inc(x int) int {
	/*@ assert x == 2; */
	x = x + 1
	return x
}
