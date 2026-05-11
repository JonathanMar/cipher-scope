package core

import (
	"context"
	"sync"
)

func RunEngine(ctx context.Context, jobChan <-chan Job, workers int) string {
	resultChan := make(chan string, 1)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobChan:
					if !ok {
						return
					}

					found, result := Process(job)
					if found {
						select {
						case resultChan <- result:
							cancel()
						default:
						}
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		return res
	}

	return ""
}