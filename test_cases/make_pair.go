package acsl

type pair struct {
	first int
	second int
}

/*@
assigns \nothing;
ensures result: \result.first == first;
ensures result: \result.second == second;
*/
func make_pair(first int, second int) pair {
	var pair pair
	pair.first = first
	pair.second = second
	return pair
}
