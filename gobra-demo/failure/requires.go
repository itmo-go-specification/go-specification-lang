package failure

/*
@
requires x >= 0;
ensures res == x + 1;
assigns \nothing;
*/
func AddOneNegativeSimpleFailed(x int) (res int) {
	return x + 1
}

func main() {
	AddOneNegativeSimpleFailed(-1)
}
