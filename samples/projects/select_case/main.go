package main

func main() {
	ch := make(chan int, 1)
	ch <- 1
	select {
	case <-ch:
		/*@ assert \false; */
		return
	}
}
