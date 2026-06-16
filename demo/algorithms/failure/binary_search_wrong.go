package failure

/*@
requires \forall integer i; (0 <= i) && (i < len(array) - 1) ==> (array[i] <= array[i+1]);
ensures (ok == 1) ==> (array[idx] == target);
ensures (ok == 1) || (ok == 0);
assigns \nothing;
*/
func BinarySearchWrong(array []int, target int) (idx int, ok int) {
	// Intentionally wrong: returns index 0 when target exists at any position.
	for i := 0; i < len(array); i++ {
		if array[i] == target {
			return 0, 1
		}
	}
	return -1, 0
}
