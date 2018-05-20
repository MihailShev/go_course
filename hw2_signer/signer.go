package main

import (
	"sort"
	"strings"
	"sync"
	"strconv"
	"fmt"
)

type multiHashResult struct {
	index int
	hash  string
}

func SingleHash(in, out chan interface{}) {
	tempCh := make(chan string)
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		convertedData := fmt.Sprintf("%v", data)
		hashMd5 := DataSignerMd5(convertedData)

		go func(tempCh chan string) {
			defer wg.Done()

			crc32ch := make(chan string)
			md5ch := make(chan string)

			go func(ch chan<- string) {
				ch <- DataSignerCrc32(convertedData)
			}(crc32ch)

			go func(ch chan<- string) {
				ch <- DataSignerCrc32(hashMd5)
			}(md5ch)

			result := <-crc32ch + "~" + <-md5ch
			tempCh <- result
		}(tempCh)

	}

	go func() {
		wg.Wait()
		close(tempCh)
	}()

	for res := range tempCh {
		out <- res
	}
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	tempCh := make(chan string)

	for data := range in {
		wg.Add(1)
		go getMultiHashResult(data, tempCh, wg)
	}

	go func() {
		wg.Wait()
		close(tempCh)
	}()

	for res := range tempCh {
		out <- res
	}
}

func getMultiHashResult(data interface{}, resultCh chan string, inWg *sync.WaitGroup ) {
	defer inWg.Done()

	temp := make([]string, 6)
	wg := &sync.WaitGroup{}
	hashResult := make(chan multiHashResult)

	for i := 0; i < 6; i++ {
		th := strconv.Itoa(i)
		wg.Add(1)
		go func(i int) {
			hashResult <- multiHashResult{index: i, hash: DataSignerCrc32(th + data.(string))}
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(hashResult)
	}()

	for res := range hashResult {
		temp[res.index] = res.hash
	}

	resultCh <- strings.Join(temp, "")
}


func CombineResults(in, out chan interface{}) {
	var resultList []string

	for result := range in {
		resultList = append(resultList, result.(string))
	}

	sort.Strings(resultList)
	out <- strings.Join(resultList, "_")
}

func ExecutePipeline(hashSignJobs ...job) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	var in chan interface{}

	for _, jobItem := range hashSignJobs {
		wg.Add(1)
		out := make(chan interface{})
		go func(jobFunc job, in chan interface{}) {
			defer close(out)
			defer wg.Done()
			jobFunc(in, out)
		}(jobItem, in)
		in = out
	}
}
