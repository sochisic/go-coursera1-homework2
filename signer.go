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

type chPack struct {
	in    chan interface{}
	out   chan interface{}
	jobIn chan interface{}
}

type chSlice []chPack

// var chans [5]chan interface{}

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
	// wg.Add(len(jobs))
	counter := 0

	for i, jb := range jobs {
		i := i

		if i == 0 {
			wg.Add(1)
			// fmt.Printf("%v Iteration\n", i)

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

			jb(chans[i].in, chans[i].out)
		}

		if i != 0 && i != len(jobs)-1 {
			wg.Add(1)
			// fmt.Printf("%v Iteration\n", i)
			go jb(chans[i].jobIn, chans[i].out)

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
			// fmt.Printf("%v Iteration\n", i)
			go jb(chans[i].jobIn, chans[i].out)

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
						time.Sleep(time.Millisecond * 10)
						return
					}
				}
			}(chans[i].jobIn, chans[i].in, chans[i].out, &wg)
		}
	}

	wg.Wait()
}

// func ExecutePipeline(jobs ...job) {
// 	var wg sync.WaitGroup
// 	chans := chanMaker(len(jobs))
// 	wg.Add(len(jobs))
// 	fmt.Println("chans", chans)

// 	for i, jb := range jobs {
// 		i := i
// 		go jb(chans[i].in, chans[i].out)
// 		// time.Sleep(10 * time.Millisecond)

// 		go func(jb job, in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
// 			defer wg.Done()
// 			if i == len(jobs)-1 {
// 			LOOP:
// 				for {
// 					// time.Sleep(10 * time.Millisecond)

// 					select {
// 					case val, ok := <-out:
// 						if !ok {
// 							break LOOP
// 						}
// 						in <- val
// 					default:
// 						fmt.Println("select default", i)
// 						break LOOP
// 					}
// 				}
// 			} else {
// 			LOOP2:
// 				for {
// 					time.Sleep(10 * time.Millisecond)

// 					select {
// 					case val, ok := <-out:
// 						if !ok {
// 							break LOOP2
// 						}
// 						chans[i+1].in <- val

// 					default:
// 						fmt.Println("select default", i)
// 						break LOOP2
// 					}
// 				}
// 			}

// 			// for d := range out {

// 			// 	chans[i+1].in <- d
// 			// 	fmt.Println("Sent", i)

// 			// }

// 			fmt.Println("Out Closed")
// 			close(in)
// 			close(out)
// 			return
// 		}(jb, chans[i].in, chans[i].out, &wg)

// 	}

// 	wg.Wait()
// }

// func worker(jb job, done, in, out chan interface{}) {
// 	jb(in, out)

// 	for n := range out {
// 		fmt.Println("слушаем out", n)
// 	}
// }

// func worker(jb job, in, out chan interface{}, wg sync.WaitGroup, i int) {
// 	defer wg.Done()
// 	ch := make(chan interface{}, 1)
// 	fmt.Printf("Worker %v Started\n", i)
// 	go jb(ch, out)
// 	for {
// 		select {
// 		case data := <-out:
// 			fmt.Println("Data in chan!")
// 			ch <- data
// 		case <-time.After(time.Second * 3):
// 			fmt.Printf("Worker %v timeout\n", i)
// 			close(in)
// 			close(ch)
// 			return
// 		}
// 	}
// }

// func ExecutePipeline(jobs ...job) {
// 	type chanSlice []chan interface{}
// 	var wg sync.WaitGroup
// 	channels := make(chanSlice, len(jobs))
// 	for i, job := range jobs {
// 		fmt.Println("i", i, jobs, len(jobs))
// 		channels[i] = make(chan interface{}, 1)
// 		wg.Add(len(jobs))

// 		if i == 0 {
// 			go worker(job, channels[i], channels[i], wg, i)
// 			// go job(channels[i], channels[i])
// 			continue
// 		}

// 		go worker(job, channels[i-1], channels[i], wg, i)
// 		// go job(channels[i-1], channels[i])
// 		fmt.Println("jobs", i, job, len(jobs), channels)
// 	}
// 	wg.Wait()
// }

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
