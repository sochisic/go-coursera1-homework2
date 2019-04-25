package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

var md5Quote = 1
var finalResult string

// сюда писать код
func main() {
	// md5QuoteChan := make(chan struct{}, md5Quote)
	// wg := &sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})
	// in2 := make(chan interface{})
	out2 := make(chan interface{})
	go SingleHash(in, out)
	in <- 1
	close(in)

	// fmt.Println("DataSignerCrc32", DataSignerCrc32(strconv.Itoa(0)))
	// for r := range out {
	// 	fmt.Println(r)
	// }

	go MultiHash(out, out2)

	res := <-out2

	fmt.Println("res", res)
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
	var counter uint64

	for i, jb := range jobs {
		if i != len(jobs)-1 {
			go jb(chans[i].jobIn, chans[i].out)
		}
	}

	// for i, jb := range jobs {
	// 	if i == 0 || i == len(jobs)-1 || i == len(jobs)-2 {
	// 		go jb(chans[i].jobIn, chans[i].out)
	// 	}
	// }

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
						atomic.AddUint64(&counter, 1)
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
				var cnt uint64

				for data := range in {
					fmt.Printf("Worker%v got data from jb%v: %v\n", i, i-1, data)
					// go jb(chans[i].jobIn, chans[i].out)
					jobIn <- data
				}
				close(jobIn)

				for data := range chans[i].out {
					fmt.Printf("Worker%v Recieved data from Job%v: %v\n", i, i, data)
					chans[i+1].in <- data
					atomic.AddUint64(&cnt, 1)

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
				var cnt uint64

				for data := range in {
					fmt.Println("Worker Last: ", data)
					jobIn <- data
					atomic.AddUint64(&cnt, 1)

					fmt.Println("counter", cnt, counter)

					if len(jobs) > 3 && cnt == 1 {
						close(jobIn)
						fmt.Println("Worker Last: Ended")
						return
					}
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
		md5 := DataSignerMd5(strconv.Itoa(data))

		ds2result := DataSignerCrc32(md5)
		ds1result := <-ds1ch

		result := ds1result + "~" + ds2result
		// fmt.Println("SingleHash", result)
		out <- result
		// close(out)
	}
	fmt.Println("SingleHash ended")
	return
}

func DataSigner(data string) chan string {
	// fmt.Println("DataSigner", data)
	resultCh := make(chan string, 1)
	go func(ch chan string) {
		d := DataSignerCrc32(data)
		ch <- d
	}(resultCh)
	// fmt.Println("DataSigner ended")
	return resultCh
}

func MultiHash(in, out chan interface{}) {
	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			errors.New("Not a string")
		}
		partCh1 := DataSigner(strconv.Itoa(0) + data)
		partCh2 := DataSigner(strconv.Itoa(1) + data)
		partCh3 := DataSigner(strconv.Itoa(2) + data)
		partCh4 := DataSigner(strconv.Itoa(3) + data)
		partCh5 := DataSigner(strconv.Itoa(4) + data)
		partCh6 := DataSigner(strconv.Itoa(5) + data)

		part1, part2, part3, part4, part5, part6 := <-partCh1, <-partCh2, <-partCh3, <-partCh4, <-partCh5, <-partCh6

		result := part1 + part2 + part3 + part4 + part5 + part6
		// fmt.Println("result", result)
		out <- result
	}
}

func CombineResults(in, out chan interface{}) {
	var result string
	dataSlice := make([]string, 0)

	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			errors.New("Not a string")
		}
		fmt.Println("CombineResults1", data)

		dataSlice = append(dataSlice, data)
	}

	sort.Strings(dataSlice)

	for i, v := range dataSlice {
		if i == 0 {
			result = v
			continue
		}
		result = result + "_" + v
	}

	fmt.Println("CombineResults2 dataSlice len", len(dataSlice))
	fmt.Println("CombineResults2", result)
	out <- result
	close(out)
}
