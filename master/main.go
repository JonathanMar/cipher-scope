package main

import (
	"auditor/core"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const address = ":9000"

func main() {

	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Master listening on port 9000")

	var (
		workers []net.Conn
		mu      sync.Mutex
	)

	// ===== ACCEPT WORKERS =====

	go func() {

		for {

			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			mu.Lock()
			workers = append(workers, conn)
			mu.Unlock()

			fmt.Println("Worker connected:", conn.RemoteAddr())
		}
	}()

	// ===== INPUT =====

	var (
		targetHash string
		hashType   string
		length     int
	)

	fmt.Print("Enter hash: ")
	fmt.Scanln(&targetHash)

	fmt.Print("Enter hash type (md5, sha1, sha256): ")
	fmt.Scanln(&hashType)

	fmt.Print("Password length: ")
	fmt.Scanln(&length)

	if err := core.ValidateHash(hashType, targetHash); err != nil {
		fmt.Println("Validation error:", err)
		return
	}

	if length < 2 {
		fmt.Println("Minimum password length is 2")
		return
	}

	var expectedWorkers int

	fmt.Print("How many workers do you want to wait for? ")
	fmt.Scanln(&expectedWorkers)

	// ===== CHARSET =====

	// Para testes rápidos:
	// charset := "abcdefghijklmnopqrstuvwxyz"

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	total := uint64(
		math.Pow(
			float64(len(charset)),
			float64(length),
		),
	)

	// ===== WAIT WORKERS =====

	fmt.Println("Waiting for workers...")

	for {

		mu.Lock()
		current := len(workers)
		mu.Unlock()

		fmt.Printf(
			"Connected workers: %d/%d\r",
			current,
			expectedWorkers,
		)

		if current >= expectedWorkers {
			fmt.Println("\nAll workers connected.")
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	mu.Lock()
	totalWorkers := len(workers)
	mu.Unlock()

	fmt.Printf(
		"Starting distributed brute force with %d workers\n",
		totalWorkers,
	)

	// ===== RESULT CHANNEL =====

	resultChan := make(chan string, 1)

	// ===== GLOBAL PROGRESS =====

	var globalProgress atomic.Uint64

	// ===== LISTEN WORKERS =====

	mu.Lock()

	for _, worker := range workers {

		go func(conn net.Conn) {

			decoder := json.NewDecoder(conn)

			var lastProgress uint64

			for {

				var msg core.Message

				if err := decoder.Decode(&msg); err != nil {
					return
				}

				switch msg.Type {

				case "progress":

					delta := msg.Progress - lastProgress
					lastProgress = msg.Progress

					globalProgress.Add(delta)

				case "result":

					select {
					case resultChan <- msg.Result:
					default:
					}

					return
				}
			}

		}(worker)
	}

	mu.Unlock()

	// ===== DISTRIBUTE TASKS =====

	fmt.Println("Distributing tasks...")

	taskCount := 0

	mu.Lock()

	for _, c1 := range charset {

		for _, c2 := range charset {

			prefix := string(c1) + string(c2)

			task := core.Task{
				Prefix:     prefix,
				Remaining:  length - 2,
				Charset:    charset,
				TargetHash: targetHash,
				HashType:   hashType,
			}

			worker := workers[taskCount%len(workers)]

			encoder := json.NewEncoder(worker)

			if err := encoder.Encode(task); err != nil {

				fmt.Println(
					"Failed to send task:",
					err,
				)

				continue
			}

			taskCount++

			// Evita flood de socket
			time.Sleep(2 * time.Millisecond)
		}
	}

	mu.Unlock()

	fmt.Printf(
		"Distributed %d tasks\n",
		taskCount,
	)

	// ===== START TIMER =====

	start := time.Now()

	// ===== PROGRESS =====

	doneProgress := make(chan struct{})

	go func() {

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {

			select {

			case <-doneProgress:
				return

			case <-ticker.C:

				current := globalProgress.Load()

				percent := float64(current) / float64(total) * 100

				hashRate := float64(current) / time.Since(start).Seconds()

				var eta time.Duration

				if hashRate > 0 {

					remaining := float64(total-current) / hashRate

					eta = time.Duration(remaining) * time.Second
				}

				const barSize = 30

				filled := int(
					(percent / 100) * float64(barSize),
				)

				bar := ""

				for i := 0; i < filled; i++ {
					bar += "="
				}

				for i := filled; i < barSize; i++ {
					bar += " "
				}

				fmt.Printf(
					"\r[%s] %.6f%% | %d/%d | %.0f H/s | ETA: %s",
					bar,
					percent,
					current,
					total,
					hashRate,
					eta.Truncate(time.Second),
				)
			}
		}
	}()

	// ===== WAIT RESULT =====

	result := <-resultChan

	close(doneProgress)

	elapsed := time.Since(start).Truncate(
		time.Millisecond,
	)

	fmt.Println()

	if result != "" {
		fmt.Println("Password found:", result)
	} else {
		fmt.Println("Password not found")
	}

	fmt.Println("Elapsed time:", elapsed)

	// ===== CLOSE WORKERS =====

	mu.Lock()

	for _, worker := range workers {
		worker.Close()
	}

	mu.Unlock()

	fmt.Println("All workers stopped.")
}
