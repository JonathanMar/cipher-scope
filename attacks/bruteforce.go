package attacks

import (
	"auditor/core"
	"context"
	"runtime"
)

func BruteForceAttack(charset string, length int, targetHash string, hashType string) string {
	bufferSize := runtime.NumCPU() * 100
	jobChan := make(chan core.Job, bufferSize)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(jobChan)

		var generate func([]byte, int)

		generate = func(prefix []byte, remaining int) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if remaining == 0 {
				jobChan <- core.Job{
					Word:       string(prefix),
					TargetHash: targetHash,
					HashType:   hashType,
				}
				return
			}

			for i := 0; i < len(charset); i++ {
				next := make([]byte, len(prefix)+1)
				copy(next, prefix)
				next[len(prefix)] = charset[i]

				generate(next, remaining-1)
			}
		}

		generate([]byte{}, length)
	}()

	return core.RunEngine(ctx, jobChan, runtime.NumCPU())
}