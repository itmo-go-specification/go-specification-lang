package math

/*@
requires (n >= 0) && (n <= 12);
ensures res >= 1;
assigns \nothing;
*/
func Factorial(n int) (res int) {
	res = 1
	for i := 2; i <= n; i++ {
		res *= i
	}
	return res
}
