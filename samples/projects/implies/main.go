package main

var zero int

func main() {
	/*@ assert \false ==> \true; */
	/*@ assert \false ==> (1 / zero == 1); */
	return
}
