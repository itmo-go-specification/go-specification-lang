package main

/*
@
assigns dest[0..n-1];
ensures n >= 0;
*/
func CopyArray(src []int, dest []int, n int) []int {
	for i := 0; i < n; i++ {
		dest[i] = src[i]
	}
	return dest
}
