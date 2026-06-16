package order

/*@
requires unitCents >= 0;
ensures res == unitCents;
assigns \nothing;
*/
func UnitPrice(unitCents int) (res int) {
	return unitCents
}

/*@
requires totalCents >= 0;
requires (percent >= 0) && (percent <= 100);
ensures res >= 0;
ensures res <= totalCents;
assigns \nothing;
*/
func ApplyDiscount(totalCents, percent int) (res int) {
	discount := totalCents * percent / 100
	return totalCents - discount
}

/*@
requires minQty <= maxQty;
ensures res >= minQty;
ensures res <= maxQty;
assigns \nothing;
*/
func ClampQuantity(qty, minQty, maxQty int) (res int) {
	if qty < minQty {
		return minQty
	}
	if qty > maxQty {
		return maxQty
	}
	return qty
}

/*@
requires totalCents >= 0;
requires parts > 0;
ensures perPart >= 0;
ensures (remainder >= 0) && (remainder < parts);
ensures perPart * parts + remainder == totalCents;
assigns \nothing;
*/
func SplitEvenly(totalCents, parts int) (perPart int, remainder int) {
	perPart = totalCents / parts
	remainder = totalCents % parts
	return perPart, remainder
}

/*@
requires unitCents >= 0;
requires qty >= 0;
ensures res == unitCents * qty;
assigns \nothing;
*/
func OrderTotal(unitCents, qty int) (res int) {
	return unitCents * qty
}
