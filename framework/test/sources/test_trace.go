package sources

import "fmt"

func helper() string {
	return "helper result"
}

func process(data string) string {
	result := helper()
	return data + " " + result
}

func calculate(a, b int) int {
	return a + b
}

func main() {
	fmt.Println("Starting program")
	
	result := process("test")
	fmt.Println(result)
	
	sum := calculate(5, 10)
	fmt.Printf("Sum: %d\n", sum)
	
	fmt.Println("Program finished")
}