package main

/*@
assigns arr[0..n-1];
ensures n >= 0;
*/
func Reverse(arr []int, n int) []int {
	left := 0
	right := n - 1

	for left < right {
		swapArrayElements(arr, left, right)
		left++
		right--
	}

	return arr
}