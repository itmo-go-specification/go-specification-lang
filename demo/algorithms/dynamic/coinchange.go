package dynamic

/*@
requires amount >= 0;
ensures res >= 0;
assigns \nothing;
*/
func CoinChange(coins []int, amount int) (res int) {
	if amount == 0 {
		return 0
	}
	dp := make([]int, amount+1)
	for i := 1; i <= amount; i++ {
		dp[i] = amount + 1
	}
	for i := 1; i <= amount; i++ {
		for _, c := range coins {
			if c <= i && dp[i-c]+1 < dp[i] {
				dp[i] = dp[i-c] + 1
			}
		}
	}
	if dp[amount] > amount {
		return -1
	}
	return dp[amount]
}
