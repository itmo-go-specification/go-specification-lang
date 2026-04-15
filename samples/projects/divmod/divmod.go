package divmod

/*
@
requires b != 0;
ensures \result_0 == a / b;
ensures \result_1 == a % b;
ensures \result_0 * b + \result_1 == a;
*/
func DivMod(a, b int) (int, int) {
	return a / b, a % b
}
