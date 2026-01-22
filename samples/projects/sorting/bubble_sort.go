package main

/*@
assigns arr[0..n-1];
ensures n >= 0;
*/
func BubbleSort(arr []int, n int) []int {
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if arr[j] > arr[j+1] {
				swapArrayElements(arr, j, j+1)
			}
		}
	}
	return arr
}