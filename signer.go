package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
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
	close(in)

	fmt.Println("DataSignerCrc32", DataSignerCrc32(strconv.Itoa(0)))
	for r := range out {
		fmt.Println(r)
	}

	// res2 := MultiHash(res)
	// fmt.Println(res2)
	// res3 := CombineResult(res2)
	// fmt.Println(res3)
}

type chPack struct {
	in    chan interface{}
	out   chan interface{}
	jobIn chan interface{}
}

type chSlice []chPack

func chanMaker(n int) (list chSlice) {
	for i := 0; i < n; i++ {
		var tmp chPack
		tmp.in = make(chan interface{}, MaxInputDataLen)
		tmp.out = make(chan interface{}, MaxInputDataLen)
		tmp.jobIn = make(chan interface{}, MaxInputDataLen)
		list = append(list, tmp)
	}
	return
}

func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup
	chans := chanMaker(len(jobs))
	counter := 0

	for i, jb := range jobs {
		if i != len(jobs)-1 {
			go jb(chans[i].jobIn, chans[i].out)
		}
	}

	for i, jb := range jobs {
		i := i

		if i == 0 {
			wg.Add(1)

			go func(in, out chan interface{}, wg *sync.WaitGroup) {
				defer wg.Done()

				for {
					select {
					case data := <-out:
						fmt.Println("Worker1 listener: ", data)
						counter++
						chans[i+1].in <- data
					default:
						fmt.Println("Worker1: Ended")
						close(chans[i+1].in)
						return
					}
				}
			}(chans[i].in, chans[i].out, &wg)
		}

		if i != 0 && i != len(jobs)-1 {
			wg.Add(1)

			go func(jobIn, in, out chan interface{}, wg *sync.WaitGroup) {
				defer wg.Done()
				cnt := 0

				for data := range in {
					fmt.Printf("Worker%v got data from jb%v: %v\n", i, i-1, data)
					jobIn <- data
				}
				close(jobIn)

				for data := range chans[i].out {
					fmt.Printf("Worker%v Recieved data from Job%v: %v\n", i, i, data)
					chans[i+1].in <- data
					cnt++

					if cnt == counter {
						close(chans[i+1].in)
						fmt.Printf("Worker%v: Ended\n", i)
						return
					}
				}
			}(chans[i].jobIn, chans[i].in, chans[i].out, &wg)
		}

		if i == len(jobs)-1 {
			wg.Add(1)

			go func(jobIn, in, out chan interface{}, wg *sync.WaitGroup) {
				defer wg.Done()
				cnt := 0

				for data := range in {
					fmt.Println("Worker Last: ", data)
					jobIn <- data
					cnt++

					if cnt == counter {
						close(jobIn)
						fmt.Println("Worker Last: Ended")
						return
					}
				}
			}(chans[i].jobIn, chans[i].in, chans[i].out, &wg)

			jb(chans[i].jobIn, chans[i].out)
		}
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	for dataRaw := range in {
		data, ok := dataRaw.(int)
		if !ok {
			errors.New("Not a Int")
		}

		ds1ch := DataSigner(strconv.Itoa(data))

		ds2result := DataSignerCrc32(DataSignerMd5(strconv.Itoa(data)))
		ds1result := <-ds1ch

		result := ds1result + "~" + ds2result
		fmt.Println("SingleHash", result)
		out <- result
		// close(out)
	}
	fmt.Println("SingleHash ended")
	return
}

func DataSigner(data string) chan string {
	fmt.Println("DataSigner", data)
	resultCh := make(chan string, 1)
	go func(ch chan string) {
		d := DataSignerCrc32(data)
		ch <- d
	}(resultCh)
	fmt.Println("DataSigner ended")
	return resultCh
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
