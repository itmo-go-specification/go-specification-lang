package math

/*@
requires n >= 0;
ensures res >= 0;
assigns \nothing;
*/
func Fibonacci(n int) (res int) {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

/*@
requires n >= 0;
ensures res >= 0;
assigns \nothing;
*/
func FibonacciIterative(n int) (res int) {
	if n <= 1 {
		return n
	}
	prev, curr := 0, 1
	for i := 2; i <= n; i++ {
		prev, curr = curr, prev+curr
	}
	return curr
}
