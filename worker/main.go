package main

import (
	"auditor/core"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"runtime"
)

func main() {
	conn, err := net.Dial("tcp", "192.168.0.66:9000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connected to master")

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var task core.Task

		err := decoder.Decode(&task)
		if err != nil {
			fmt.Println("Connection closed:", err)
			return
		}

		fmt.Println("Received task with prefix:", task.Prefix)

		result := runTask(task)

		if result != "" {
			fmt.Println("Found password:", result)
			encoder.Encode(result)
			return
		}
	}
}

func runTask(task core.Task) string {
	jobChan := make(chan core.Job, runtime.NumCPU()*100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(jobChan)

		var generate func(string, int)

		generate = func(prefix string, remaining int) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if remaining == 0 {
				jobChan <- core.Job{
					Word:       prefix,
					TargetHash: task.TargetHash,
					HashType:   task.HashType,
				}
				return
			}

			for _, c := range task.Charset {
				generate(prefix+string(c), remaining-1)
			}
		}

		generate(task.Prefix, task.Remaining)
	}()

	return core.RunEngine(ctx, jobChan, runtime.NumCPU())
}