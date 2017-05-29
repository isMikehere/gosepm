package test

import (
	"fmt"
	"time"
)

func copySlice() {
	sl_from := []int{1, 2, 3}
	sl_to := make([]int, 1, 3)

	n := copy(sl_to, sl_from)
	// var sl_to []int = sl_from[1:10]
	// fmt.Println(sl_to)
	fmt.Printf("Copied %d elements\n", n) // n == 3

	// sl3 := []int{1, 2, 3}
	// sl3 = append(sl3, 4, 5, 6, 7)
	sl_to = append(sl_to, 4, 5, 6)
	fmt.Println(sl_to)
}
func TestChan() {
	ch := make(chan string)

	go sendData(ch)
	// go getData(ch)

	go func() {
		input := <-ch
		fmt.Println(input)
	}()

	time.Sleep(1e9)
}

func sendData(ch chan string) {
	ch <- "Washington"
}

func getData(ch chan string) {
	var input string
	// time.Sleep(2e9)
	for {
		input = <-ch
		fmt.Printf("%s ", input)
	}
}
