package main

import (
	"fmt"
)

func printArray(arr []int, name string) {
	fmt.Printf("%s: [", name)
	for i := 0; i < len(arr); i++ {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", arr[i])
	}
	fmt.Printf("]\n")
}

func main() {
	arr := []int{64, 34, 25, 12, 22, 11, 90}
	n := len(arr)

	fmt.Println("=== Array Operations Demo ===\n")

	printArray(arr, "Original array")
	fmt.Println()

	minIdx := FindMin(arr, n)
	fmt.Printf("Minimum element index: %d (value: %d)\n\n", minIdx, arr[minIdx])

	copied := make([]int, n)
	CopyArray(arr, copied, n)
	printArray(copied, "Copied array")
	fmt.Println()

	Reverse(copied, n)
	printArray(copied, "Reversed array")
	fmt.Println()

	sorted := make([]int, n)
	CopyArray(arr, sorted, n)
	BubbleSort(sorted, n)
	printArray(sorted, "Sorted array")
	fmt.Println()

	printArray(arr, "Original (unchanged)")
}
