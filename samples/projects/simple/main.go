package main

import (
	"fmt"
)

/*@
requires (f == \true);
assigns \nothing;
*/
func check(f bool) {
    fmt.Printf("Passed\n")
}

func main() {
    check(false);
}
