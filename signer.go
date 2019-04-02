package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var md5Quote = 1
var finalResult string

// сюда писать код
func main() {
	// md5QuoteChan := make(chan struct{}, md5Quote)
	// wg := &sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})
	go SingleHash(in, out)
	in <- 1

	for r := range out {
		fmt.Println(r)
	}
	// res2 := MultiHash(res)
	// fmt.Println(res2)
	// res3 := CombineResult(res2)
	// fmt.Println(res3)
}

func worker(jb job, in, out chan interface{}, wg sync.WaitGroup, i int) {
	defer wg.Done()
	ch := make(chan interface{}, 1)
	fmt.Printf("Worker %v Started\n", i)
	go jb(ch, out)
	for {
		select {
		case data := <-out:
			fmt.Println("Data in chan!")
			ch <- data
		case <-time.After(time.Second * 3):
			fmt.Printf("Worker %v timeout\n", i)
			close(in)
			close(ch)
			return
		}
	}
}

func ExecutePipeline(jobs ...job) {
	type chanSlice []chan interface{}
	var wg sync.WaitGroup
	channels := make(chanSlice, len(jobs))
	for i, job := range jobs {
		fmt.Println("i", i, jobs, len(jobs))
		channels[i] = make(chan interface{}, 1)
		wg.Add(len(jobs))

		if i == 0 {
			go worker(job, channels[i], channels[i], wg, i)
			// go job(channels[i], channels[i])
			continue
		}

		go worker(job, channels[i-1], channels[i], wg, i)
		// go job(channels[i-1], channels[i])
		fmt.Println("jobs", i, job, len(jobs), channels)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			errors.New("Not a string")
		}

		result := DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
		out <- result
	}
	// return out
}

func MultiHash(in, out chan interface{}) {
	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			errors.New("Not a string")
		}
		var result string
		for i := 0; i < 6; i++ {
			part := DataSignerCrc32(strconv.Itoa(i) + data)
			result += part
		}
		out <- result
	}
}

func CombineResults(in, out chan interface{}) {
	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			errors.New("Not a string")
		}

		if len([]byte(finalResult)) == 0 {
			finalResult = data
			out <- data
		} else {
			var result string
			result = finalResult + "_" + data
			out <- result
		}
	}
}

// func Crc32Worker(i int, result chan) {

// }
