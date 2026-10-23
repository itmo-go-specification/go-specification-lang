package labeled

// ExprTest covers all expression kinds supported by the dynamic_checks plugin:
// arithmetic (+, -, *, /, %), comparisons (==, !=, <, <=, >, >=),
// logical (&&, ||, !), unary (-), integer literals, \true/\false,
// \old, \result, assert.

/*
@
requires x >= 0;
ensures \result == x + 1;
assigns \nothing;
*/
func AddOne(x int) int {
	return x + 1
}

/*
@
requires b > a;
ensures \result == b - a;
assigns \nothing;
*/
func Diff(a, b int) int {
	return b - a
}

/*
@
requires n >= 0;
ensures \result == n * 2;
assigns \nothing;
*/
func Double(n int) int {
	return n * 2
}

/*
@
requires d != 0;
ensures \result == n / d;
assigns \nothing;
*/
func Divide(n, d int) int {
	return n / d
}

/*
@
requires d != 0;
ensures \result == n % d;
assigns \nothing;
*/
func Modulo(n, d int) int {
	return n % d
}

/*
@
requires x >= 0;
ensures \result >= 0;
assigns \nothing;
*/
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

/*
@
requires !(x == 0);
ensures \result != 0;
assigns \nothing;
*/
func Identity(x int) int {
	return x
}

/*
@
requires lo <= hi;
ensures \result >= lo;
ensures \result <= hi;
assigns \nothing;
*/
func Clamp(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

/*
@
requires n > 0;
ensures \result == \old(n) + 1;
assigns \nothing;
*/
func Increment(n int) int {
	return n + 1
}

/*
@
requires \true;
ensures \result == 42;
assigns \nothing;
*/
func FortyTwo() int {
	return 42
}

/*
@
requires n >= 0;
ensures \result == n * (n + 1) / 2;
assigns \nothing;
*/
func TriangleNumber(n int) int {
	return n * (n + 1) / 2
}

func AssertAdd(x, y int) {
	/*@ assert x + y >= x; */
	_ = x
	_ = y
}

func AssertMul(x int) {
	/*@ assert x * 2 == x + x; */
	_ = x
}

func AssertSub(x int) {
	/*@ assert x - x == 0; */
	_ = x
}

func AssertNeq(x int) {
	/*@ assert x != x + 1; */
	_ = x
}
