package notation

/*@
assigns p, q;
ensures (*p == \old(*q)) && (*q == \old(*p));
*/
func swap(p *int, q *int) {
	save := *p;
	*p = *q;
	*q = save;
}
