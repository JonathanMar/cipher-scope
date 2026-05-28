package attacks

import (
	"auditor/core"
	"bufio"
	"context"
	"os"
	"runtime"
	"strings"
)

// DictionaryAttack tenta cada palavra do arquivo wordlist contra o hash alvo.
// Usa streaming linha a linha para preservar RAM em dispositivos limitados.
func DictionaryAttack(
	path string,
	targetHash string,
	hashType string,
	onProgress func(),
) string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numWorkers := runtime.NumCPU()
	jobChan := make(chan core.Job, numWorkers*4)

	go func() {
		defer close(jobChan)

		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			if word == "" {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case jobChan <- core.Job{
				Word:       word,
				TargetHash: targetHash,
				HashType:   hashType,
			}:
			}
		}
	}()

	return core.RunEngine(ctx, jobChan, numWorkers, onProgress)
}