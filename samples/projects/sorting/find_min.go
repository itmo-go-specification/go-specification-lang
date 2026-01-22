package main

/*@
assigns \nothing;
ensures \result >= 0;
ensures \result < n;
*/
func FindMin(arr []int, n int) int {
	minIdx := 0
	for i := 1; i < n; i++ {
		if arr[i] < arr[minIdx] {
			minIdx = i
		}
	}
	return minIdx
}