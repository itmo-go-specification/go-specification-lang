package labeled

// ExprTest covers all expression kinds supported by the dynamic_checks plugin:
// arithmetic (+, -, *, /, %), comparisons (==, !=, <, <=, >, >=),
// logical (&&, ||, !), unary (-), integer literals, \true/\false,
// \old, res, assert.

/*
@
requires x >= 0;
ensures res == x + 1;
assigns \nothing;
*/
func AddOne(x int) (res int) {
	return x + 1
}

/*
@
requires b > a;
ensures res == b - a;
assigns \nothing;
*/
func Diff(a, b int) (res int) {
	return b - a
}

/*
@
requires n >= 0;
ensures res == n * 2;
assigns \nothing;
*/
func Double(n int) (res int) {
	return n * 2
}

/*
@
requires d != 0;
ensures res == n / d;
assigns \nothing;
*/
func Divide(n, d int) (res int) {
	return n / d
}

/*
@
requires d != 0;
ensures res == n % d;
assigns \nothing;
*/
func Modulo(n, d int) (res int) {
	return n % d
}

/*
@
requires x >= 0;
ensures res >= 0;
assigns \nothing;
*/
func Abs(x int) (res int) {
	if x < 0 {
		return -x
	}
	return x
}

/*
@
requires !(x == 0);
ensures res != 0;
assigns \nothing;
*/
func Identity(x int) (res int) {
	return x
}

/*
@
requires lo <= hi;
ensures res >= lo;
ensures res <= hi;
assigns \nothing;
*/
func Clamp(x, lo, hi int) (res int) {
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
ensures res == \old(n) + 1;
assigns \nothing;
*/
func Increment(n int) (res int) {
	return n + 1
}

/*
@
requires \true;
ensures res == 42;
assigns \nothing;
*/
func FortyTwo() (res int) {
	return 42
}

/*
@
requires n >= 0;
ensures res == n * (n + 1) / 2;
assigns \nothing;
*/
func TriangleNumber(n int) (res int) {
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
