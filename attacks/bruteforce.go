package attacks

import (
	"auditor/core"
	"context"
	"runtime"
)

// BruteForceAttack gera todas as combinações de exatamente `length` caracteres.
func BruteForceAttack(
	ctx context.Context,
	charset string,
	length int,
	targetHash string,
	hashType string,
	onProgress func(),
) string {
	numWorkers := runtime.NumCPU()
	jobChan := make(chan core.Job, numWorkers*4)

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

	return core.RunEngine(ctx, jobChan, numWorkers, onProgress)
}

// BruteForceAttackUpTo tenta todos os comprimentos de 1 até maxLength.
func BruteForceAttackUpTo(
	ctx context.Context,
	charset string,
	maxLength int,
	targetHash string,
	hashType string,
	onProgress func(),
) string {
	for l := 1; l <= maxLength; l++ {
		select {
		case <-ctx.Done():
			return ""
		default:
		}
		result := BruteForceAttack(ctx, charset, l, targetHash, hashType, onProgress)
		if result != "" {
			return result
		}
	}
	return ""
}
