package dynamic

/*@
requires len(nums) > 0;
ensures \true;
assigns \nothing;
*/
func MaxSubarray(nums []int) (res int) {
	res = nums[0]
	current := nums[0]
	for i := 1; i < len(nums); i++ {
		if current+nums[i] < nums[i] {
			current = nums[i]
		} else {
			current += nums[i]
		}
		if current > res {
			res = current
		}
	}
	return res
}
