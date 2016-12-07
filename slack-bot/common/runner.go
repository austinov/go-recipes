package common

import "sync"

type Worker func(in <-chan interface{}, out chan<- interface{})

// runWorkers runs several workers passing them in/out channels.
func RunWorkers(wg *sync.WaitGroup,
	in <-chan interface{}, out chan<- interface{},
	numWorkers int, w Worker) {
	go func() {
		defer func() {
			if out != nil {
				close(out)
			}
			wg.Done()
		}()

		var wg_ sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			wg_.Add(1)
			go func() {
				defer wg_.Done()
				w(in, out)
			}()
		}
		wg_.Wait()
	}()
}
