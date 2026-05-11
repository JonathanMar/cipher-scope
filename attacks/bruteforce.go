package attacks

import (
	"auditor/core"
	"context"
	"runtime"
)

func BruteForceAttack(
	charset string,
	length int,
	targetHash string,
	hashType string,
) string {

	jobChan := make(chan core.Job, runtime.NumCPU()*4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(jobChan)

		buffer := make([]byte, length)

		var generate func(int)

		generate = func(pos int) {

			select {
			case <-ctx.Done():
				return
			default:
			}

			if pos == length {

				word := string(buffer)

				select {
				case <-ctx.Done():
					return

				case jobChan <- core.Job{
					Word:       word,
					TargetHash: targetHash,
					HashType:   hashType,
				}:
				}

				return
			}

			for i := 0; i < len(charset); i++ {
				buffer[pos] = charset[i]
				generate(pos + 1)
			}
		}

		generate(0)
	}()

	return core.RunEngine(
		ctx,
		jobChan,
		runtime.NumCPU(),
		nil,
	)
}
