package main

import (
	"flag"
	"fmt"
	"os"

	dyn "gpma/demo/algorithms/dynamic"
	algfailure "gpma/demo/algorithms/failure"
	"gpma/demo/algorithms/math"
	"gpma/demo/algorithms/search"
	"gpma/demo/algorithms/sort"
	strpkg "gpma/demo/algorithms/strings"
	"gpma/demo/order"
	orderfailure "gpma/demo/order/failure"
)

func main() {
	section := flag.String("section", "all", "demo section: algorithms, order, or all")
	bad := flag.Bool("bad", false, "run failure examples with incorrect specifications")
	flag.Parse()

	switch *section {
	case "algorithms":
		runAlgorithms(*bad)
	case "order":
		runOrder(*bad)
	case "all":
		runAlgorithms(*bad)
		runOrder(*bad)
	default:
		fmt.Fprintf(os.Stderr, "unknown section %q (use algorithms, order, or all)\n", *section)
		os.Exit(1)
	}
}

func runAlgorithms(bad bool) {
	fmt.Println("=== Algorithms ===")

	fmt.Println("math:")
	fmt.Printf("  GCD(48,18)=%d Abs(-7)=%d Clamp(15,0,10)=%d Factorial(5)=%d Fibonacci(10)=%d\n",
		math.GCD(48, 18), math.Abs(-7), math.Clamp(15, 0, 10),
		math.Factorial(5), math.Fibonacci(10))

	fmt.Println("search:")
	sorted := []int{1, 3, 5, 7, 9, 11}
	idx, ok := search.BinaryIterative(sorted, 7)
	fmt.Printf("  BinaryIterative(7): idx=%d ok=%d\n", idx, ok)
	idx, ok = search.Linear(sorted, 5)
	fmt.Printf("  Linear(5): idx=%d ok=%d\n", idx, ok)

	fmt.Println("sort:")
	unsorted := []int{5, 2, 8, 1, 9}
	a := append([]int(nil), unsorted...)
	sort.InsertionSort(a)
	b := append([]int(nil), unsorted...)
	sort.BubbleSort(b)
	fmt.Printf("  InsertionSort=%v BubbleSort=%v\n", a, b)

	fmt.Println("strings:")
	fmt.Printf("  Hamming=%d IsPalindrome(racecar)=%d\n",
		strpkg.Hamming("karolin", "kathrin"), strpkg.IsPalindrome("racecar"))

	fmt.Println("dynamic:")
	fmt.Printf("  MaxSubarray=%d CoinChange=%d\n",
		dyn.MaxSubarray([]int{-2, 1, -3, 4, -1, 2, 1, -5, 4}),
		dyn.CoinChange([]int{1, 2, 5}, 5))

	if bad {
		fmt.Println("failure:")
		idx, ok = algfailure.BinarySearchWrong(sorted, 7)
		fmt.Printf("  BinarySearchWrong(7): idx=%d ok=%d\n", idx, ok)
	}
}

func runOrder(bad bool) {
	fmt.Println("=== Order ===")
	unit := order.UnitPrice(1999)
	qty := order.ClampQuantity(3, 1, 10)
	total := order.OrderTotal(unit, qty)
	discounted := order.ApplyDiscount(total, 10)
	per, rem := order.SplitEvenly(total, 3)
	fmt.Printf("  unit=%d qty=%d total=%d discounted=%d split=%d rem=%d\n",
		unit, qty, total, discounted, per, rem)

	if bad {
		wrong := orderfailure.ApplyDiscountWrong(total, 10)
		fmt.Printf("  ApplyDiscountWrong=%d\n", wrong)
	}
}
