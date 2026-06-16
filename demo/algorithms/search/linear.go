package search

/*@
requires \true;
ensures (ok == 1) || (ok == 0);
ensures (ok == 0) || ((idx >= 0) && (idx < len(array)));
assigns \nothing;
*/
func Linear(array []int, target int) (idx int, ok int) {
	for i := 0; i < len(array); i++ {
		if array[i] == target {
			return i, 1
		}
	}
	return -1, 0
}
