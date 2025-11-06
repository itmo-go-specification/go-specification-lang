package acsl

/*@
requires valid: \valid(p);
requires valid: \valid(q);
assigns *p;
assigns *q;
ensures exchange: *p == \old(*q);
ensures exchange: *q == \old(*p);
*/
func swap(p *int, q *int) {
	save := *p;
	*p = *q;
	*q = save;
}
