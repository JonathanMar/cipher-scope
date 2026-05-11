package main

import (
	"auditor/core"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

const serverAddr = "192.168.0.66:9000"

func main() {

	for {

		err := run()

		fmt.Println("Disconnected:", err)

		time.Sleep(3 * time.Second)

		fmt.Println("Reconnecting...")
	}
}

func run() error {

	conn, err := net.DialTimeout(
		"tcp",
		serverAddr,
		5*time.Second,
	)

	if err != nil {
		return err
	}

	defer conn.Close()

	fmt.Println("Connected to master")

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {

		conn.SetReadDeadline(
			time.Now().Add(30 * time.Second),
		)

		var task core.Task

		if err := decoder.Decode(&task); err != nil {
			return err
		}

		fmt.Println("Received:", task.Prefix)

		var progress atomic.Uint64

		done := make(chan string, 1)

		go func() {

			result := runTask(
				task,
				&progress,
			)

			done <- result

		}()

		ticker := time.NewTicker(1 * time.Second)

		taskFinished := false

		for !taskFinished {

			select {

			case result := <-done:

				if result != "" {

					fmt.Println(
						"Password found:",
						result,
					)

					msg := core.Message{
						Type:   "result",
						Result: result,
					}

					if err := encoder.Encode(msg); err != nil {
						ticker.Stop()
						return err
					}
				}

				taskFinished = true

			case <-ticker.C:

				msg := core.Message{
					Type:     "progress",
					Progress: progress.Load(),
				}

				if err := encoder.Encode(msg); err != nil {
					ticker.Stop()
					return err
				}
			}
		}

		ticker.Stop()
	}
}

func runTask(
	task core.Task,
	progress *atomic.Uint64,
) string {

	jobChan := make(chan core.Job, runtime.NumCPU()*4)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	go func() {

		defer close(jobChan)

		buffer := make([]byte, task.Remaining)

		var generate func(int)

		generate = func(pos int) {

			select {
			case <-ctx.Done():
				return
			default:
			}

			if pos == task.Remaining {

				word := task.Prefix + string(buffer)

				progress.Add(1)

				select {

				case <-ctx.Done():
					return

				case jobChan <- core.Job{
					Word:       word,
					TargetHash: task.TargetHash,
					HashType:   task.HashType,
				}:
				}

				return
			}

			for i := 0; i < len(task.Charset); i++ {

				buffer[pos] = task.Charset[i]

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
