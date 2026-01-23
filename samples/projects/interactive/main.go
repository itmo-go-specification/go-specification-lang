package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

/*@
requires num >= 0.0;
ensures num >= 2.0;
assigns \nothing;
*/
func foo(num float64) float64 {
	fmt.Printf("Positive number is: %v\n", num)
	num++
	/*@ assert \old(num) == (num - 1); */
	return num;
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter positive number: ")
	scanner.Scan()

	input := scanner.Text()
	num, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Printf("Error: '%s' is not number\n", input)
		os.Exit(1)
	}

	foo(num)
}
