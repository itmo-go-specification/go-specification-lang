package acsl

/*
@
requires bound: lower < upper;
assigns \nothing;
ensures bound: lower <= \result <= upper;

behavior lower_bound:

	assumes v < lower;
	assigns \nothing;
	ensures result: \result == lower;

behavior between:

	assumes lower <= v <= upper;
	assigns \nothing;
	ensures result: \result == v;

behavior upper_bound:

	assumes upper < v;
	assigns \nothing;
	ensures result: \result == upper;

complete behaviors;
disjoint behaviors;
*/
func clamp(v int, lower int, upper int) int {
	if v < lower {
		return lower
	} else if v <= upper {
		return v
	} else {
		return upper
	}
}
