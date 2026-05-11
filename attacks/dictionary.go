package attacks

import (
	"auditor/core"
	"bufio"
	"context"
	"os"
	"runtime"
)

func DictionaryAttack(
	path string,
	targetHash string,
	hashType string,
) string {

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	jobChan := make(chan core.Job, runtime.NumCPU()*4)

	go func() {

		defer close(jobChan)

		file, err := os.Open(path)
		if err != nil {
			return
		}
	