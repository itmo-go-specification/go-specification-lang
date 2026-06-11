package main

//@ predicate non_negative(x integer) = x >= 0;
//@ predicate positive(x integer) = x > 0;
//@ logic inc(x integer) integer = x + 1;

/*@
predicate lessOrEqual(x, y integer) =
    x <= y;

logic add(x, y integer) integer =
    x + y;

check lemma block_comment_checked_lemma:
    \forall x integer; x == x;
*/

/*
@
requires non_negative(n);
ensures positive(\result);
ensures \result == inc(n);
*/
func IncBlockContract(n int) int {
	return n + 1
}

func Assertions(n int) int {
	//@ assert n == n;

	/*@ assert n + 0 == n; */

	/*
		@ assert n >= 0 ==> n >= 0;
	*/

	return n
}

func LoopLineAnnotation(n int) int {
	i := 0

	//@ loop invariant 0 <= i && i <= n; loop variant n - i;
	for i < n {
		i++
	}

	return i
}

func LoopBlockAnnotation(n int) int {
	i := 0

	/*@
	loop invariant 0 <= i;
	loop invariant i <= n;
	loop variant n - i;
	*/
	for i < n {
		i++
	}

	return i
}

func MixedFormatting(n int) int {
	// simple comment. Not a spec comment

	//@ assert n == n;
	//@ assert n + 0 == n + 1 - 1;

	/*@
	assert n + 1 > n;
	*/

	/*@ assert n >= 0 ==> n + 1 > 0;
	 */

	return n
}

func main() {
	_ = IncBlockContract(1)
	_ = Assertions(1)
	_ = LoopLineAnnotation(3)
	_ = MixedFormatting(1)
}
