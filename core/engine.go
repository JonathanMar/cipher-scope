package core

import (
	"context"
	"sync"
)

func RunEngine(
	ctx context.Context,
	jobChan <-chan Job,
	workers int,
	onProgress func(),
) string {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultChan := make(chan string, 1)

	var wg sync.WaitGroup

	worker := func() {
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

				// Progresso registrado após processamento (não antes)
				if onProgress != nil {
					onProgress()
				}

				if !found {
					continue
				}

				select {
				case resultChan <- result:
					cancel()
				default:
				}

				return
			}
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		return result
	}

	return ""
}
