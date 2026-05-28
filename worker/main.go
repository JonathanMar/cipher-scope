package main

import (
	"auditor/core"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {
	masterAddr := flag.String(
		"master", "192.168.0.66:9000",
		"Endereço do servidor master (host:porta)",
	)
	numWorkers := flag.Int(
		"workers", runtime.NumCPU(),
		"Número de goroutines locais de processamento",
	)
	flag.Parse()

	fmt.Printf("Cipher-Scope Worker | master=%s | goroutines=%d\n", *masterAddr, *numWorkers)

	// Reconexão com backoff exponencial para não sobrecarregar o master
	backoff := 1 * time.Second
	const maxBackoff = 30 * time.Second

	for {
		err := run(*masterAddr, *numWorkers)

		fmt.Println("Desconectado:", err)
		fmt.Printf("Reconectando em %s...\n", backoff)

		time.Sleep(backoff)

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func run(masterAddr string, numWorkers int) error {
	conn, err := net.DialTimeout("tcp", masterAddr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Println("Conectado ao master:", masterAddr)

	// Backoff zerado após conexão bem-sucedida (feito via closure na main)
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		// Sem SetReadDeadline aqui: brute-force pode levar minutos por tarefa.
		// A detecção de queda é feita pelo erro no Decode (conn fechada pelo master).
		var task core.Task

		if err := decoder.Decode(&task); err != nil {
			return err
		}

		fmt.Printf("Tarefa recebida: prefix=%q remaining=%d\n", task.Prefix, task.Remaining)

		var progress atomic.Uint64
		done := make(chan string, 1)

		go func() {
			done <- runTask(task, &progress, numWorkers)
		}()

		ticker := time.NewTicker(1 * time.Second)

	loop:
		for {
			select {

			case result := <-done:
				ticker.Stop()

				if result != "" {
					fmt.Println("Senha encontrada:", result)

					msg := core.Message{
						Type:   "result",
						Result: result,
					}

					if err := encoder.Encode(msg); err != nil {
						return err
					}
				}

				break loop

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
	}
}

func runTask(task core.Task, progress *atomic.Uint64, numWorkers int) string {
	jobChan := make(chan core.Job, numWorkers*4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(jobChan)

		// Reusa o buffer para evitar alocações a cada palavra gerada
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

	return core.RunEngine(ctx, jobChan, numWorkers, nil)
}
