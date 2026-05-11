package attacks

import (
	"auditor/core"
	"bufio"
	"context"
	"os"
	"runtime"
)

func DictionaryAttack(path string, targetHash string, hashType string) string {
	bufferSize := runtime.NumCPU() * 100
	jobChan := make(chan core.Job, bufferSize)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(jobChan)

		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				jobChan <- core.Job{
					Word:       scanner.Text(),
					TargetHash: targetHash,
					HashType:   hashType,
				}
			}
		}
	}()

	return core.RunEngine(ctx, jobChan, runtime.NumCPU())
}