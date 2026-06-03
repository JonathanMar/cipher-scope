package attacks

import (
	"auditor/core"
	"bufio"
	"context"
	"os"
	"runtime"
	"strings"
)

func DictionaryAttack(ctx context.Context, path, targetHash, hashType string, onProgress func()) string {
	n := runtime.NumCPU()
	jobChan := make(chan core.Job, n*4)

	go func() {
		defer close(jobChan)
		f, err := os.Open(path)
		if err != nil { return }
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			word := strings.TrimSpace(sc.Text())
			if word == "" { continue }
			select {
			case <-ctx.Done(): return
			case jobChan <- core.Job{Word: word, TargetHash: targetHash, HashType: hashType}:
			}
		}
	}()

	return core.RunEngine(ctx, jobChan, n, onProgress)
}

// RulesAttack aplica mutações em cada palavra da wordlist antes de testar.
// Uma wordlist de 173k palavras gera ~12M candidatos — cobertura muito maior.
func RulesAttack(ctx context.Context, path, targetHash, hashType string, onProgress func()) string {
	n := runtime.NumCPU()
	jobChan := make(chan core.Job, n*16)

	go func() {
		defer close(jobChan)
		f, err := os.Open(path)
		if err != nil { return }
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			word := strings.TrimSpace(sc.Text())
			if word == "" { continue }

			for _, mutation := range ApplyRules(word) {
				select {
				case <-ctx.Done(): return
				case jobChan <- core.Job{Word: mutation, TargetHash: targetHash, HashType: hashType}:
				}
			}
		}
	}()

	return core.RunEngine(ctx, jobChan, n, onProgress)
}