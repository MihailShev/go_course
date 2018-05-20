package main

import (
	"fmt"
)


func main() {
	inputData := []int{0, 1}
	jobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			for res := range in {
				fmt.Println(res)
			}
		}),
	}

	ExecutePipeline(jobs...)
}
