package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
)

/*@
requires num >= 0.0;
assigns \nothing;
*/
func printNumber(num float64) {
    fmt.Printf("Positive number is: %v\n", num)
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
    
    printNumber(num)
}
