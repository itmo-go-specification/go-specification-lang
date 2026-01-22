package main

/*@
assigns a, b;
ensures a == \old(b);
ensures b == \old(a);
*/
func swapInts(a, b int) (int, int) {
	return b, a
}

/*@
assigns arr[i], arr[j];
*/
func swapArrayElements(arr []int, i, j int) {
	tempI := arr[i]
	tempJ := arr[j]
	arr[i] = tempJ
	arr[j] = tempI
}