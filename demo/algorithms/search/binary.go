package search

/*@
requires \forall integer i; (0 <= i) && (i < len(array) - 1) ==> (array[i] <= array[i+1]);
ensures (ok == 1) || (ok == 0);
assigns \nothing;
*/
func BinaryIterative(array []int, target int) (idx int, ok int) {
	start := 0
	end := len(array) - 1
	for start <= end {
		mid := start + (end-start)/2
		if array[mid] > target {
			end = mid - 1
		} else if array[mid] < target {
			start = mid + 1
		} else {
			return mid, 1
		}
	}
	return -1, 0
}

/*@
requires \forall integer i; (0 <= i) && (i < len(array) - 1) ==> (array[i] <= array[i+1]);
ensures (ok == 1) || (ok == 0);
assigns \nothing;
*/
func LowerBound(array []int, target int) (idx int, ok int) {
	start := 0
	end := len(array) - 1
	for start <= end {
		mid := start + (end-start)/2
		if array[mid] < target {
			start = mid + 1
		} else {
			end = mid - 1
		}
	}
	if start >= len(array) {
		return -1, 0
	}
	return start, 1
}
