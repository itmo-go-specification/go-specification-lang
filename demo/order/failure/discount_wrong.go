package failure

/*@
requires totalCents >= 0;
requires (percent >= 0) && (percent <= 100);
ensures res == totalCents;
assigns \nothing;
*/
func ApplyDiscountWrong(totalCents, percent int) (res int) {
	discount := totalCents * percent / 100
	return totalCents - discount
}
