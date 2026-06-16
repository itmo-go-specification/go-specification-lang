package strings

/*@
requires \true;
ensures (ok == 1) || (ok == 0);
assigns \nothing;
*/
func IsPalindrome(s string) (ok int) {
	left := 0
	right := len(s) - 1
	for left < right {
		if s[left] != s[right] {
			return 0
		}
		left++
		right--
	}
	return 1
}
