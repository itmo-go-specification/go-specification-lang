package acsl

/*@
requires 0 < n;
requires \valid_read(a + (0..n-1));
assigns \nothing;
ensures \forall integer i; 0 <= i <= n-1 ==> \result >= a[i];
ensures \exists integer e; 0 <= e <= n-1 && \result == a[e];
*/
func max_seq(a []int, n int) int {
	max := a[0]
	for i := 1; i < n; i++ {
		if a[i] > max {
			max = a[i]
		}
	}
	return max
}
