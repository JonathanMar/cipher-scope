package main

import (
	"auditor/core"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Master listening on port 9000")

	var workers []net.Conn
	var mu sync.Mutex

	// Accept workers
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
	var targetHash string
	var hashType string

	fmt.Print("Enter hash: ")
	fmt.Scanln(&targetHash)

	fmt.Print("Enter hash type (md5, sha1, sha256): ")
	fmt.Scanln(&hashType)

	charset := "abcdefghijklmnopqrstuvwxyz"
	length := 5

	// ===== WAIT FOR WORKERS =====
	fmt.Println("Waiting for at least 1 worker...")
	for {
		mu.Lock()
		if len(workers) > 0 {
			mu.Unlock()
			break
		}
		mu.Unlock()
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Starting distributed brute force with", len(workers), "workers")

	// ===== RESULT CHANNEL =====
	resultChan := make(chan string, 1)

	// Listen for results
	for _, w := range workers {
		go func(conn net.Conn) {
			decoder := json.NewDecoder(conn)

			var result string
			err := decoder.Decode(&result)
			if err == nil && result != "" {
				select {
				case resultChan <- result:
				default:
				}
			}
		}(w)
	}

	// ===== DISTRIBUTE TASKS =====
	i := 0
	for _, c := range charset {
		task := core.Task{
			Prefix:     string(c),
			Remaining:  length - 1,
			Charset:    charset,
			TargetHash: targetHash,
			HashType:   hashType,
		}

		mu.Lock()
		w := workers[i%len(workers)]
		mu.Unlock()

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(task); err != nil {
			fmt.Println("Failed to send task:", err)
		}

		i++
	}

	// ===== WAIT RESULT OR TIMEOUT =====
	select {
	case result := <-resultChan:
		fmt.Println("Password found:", result)
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout: password not found")
	}

	// ===== CLOSE WORKERS =====
	mu.Lock()
	for _, w := range workers {
		w.Close()
	}
	mu.Unlock()

	fmt.Println("All workers stopped.")
}