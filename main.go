package main

import "fmt"

func callMyName(name string) (string, error) {
	return name, nil
}

func main() {
	numberCh := make(chan int)

	//go func() {
	//	fmt.Println(<- numberCh)
	//}()
	//numberCh <- 1
	//fmt.Scanln()

	go func() {
		for i := 0; i < 5; i++ {
			fmt.Println("for:", i)
			numberCh <- i
		}
		close(numberCh)
	}()

	for num := range numberCh {
		fmt.Println(num)
	}

}
