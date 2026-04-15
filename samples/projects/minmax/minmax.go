package minmax

/*
@
requires len(a) > 0;
ensures \result_0 <= \result_1;
*/
func MinMax(a []int) (min, max int) {
	min, max = a[0], a[0]
	for _, v := range a[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}
