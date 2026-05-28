package main

import (
	"auditor/core"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	listenAddr := flag.String("addr", ":9000", "Endereço de escuta (host:porta)")
	flag.Parse()

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Master ouvindo em", *listenAddr)

	var (
		workers []net.Conn
		mu      sync.Mutex
	)

	// ===== ACEITAR WORKERS =====

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			workers = append(workers, conn)
			mu.Unlock()
			fmt.Println("Worker conectado:", conn.RemoteAddr())
		}
	}()

	// ===== INPUT =====

	var (
		targetHash      string
		hashType        string
		length          int
		expectedWorkers int
	)

	fmt.Print("Hash alvo: ")
	fmt.Scanln(&targetHash)

	fmt.Print("Tipo de hash (md5, sha1, sha256): ")
	fmt.Scanln(&hashType)

	fmt.Print("Tamanho da senha: ")
	fmt.Scanln(&length)

	if err := core.ValidateHash(hashType, targetHash); err != nil {
		fmt.Println("Erro de validação:", err)
		return
	}

	if length < 2 {
		fmt.Println("Tamanho mínimo da senha é 2")
		return
	}

	fmt.Print("Quantos workers aguardar? ")
	fmt.Scanln(&expectedWorkers)

	// ===== CHARSET =====

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	total := uint64(math.Pow(float64(len(charset)), float64(length)))

	// ===== AGUARDAR WORKERS =====

	fmt.Println("Aguardando workers...")

	for {
		mu.Lock()
		current := len(workers)
		mu.Unlock()

		fmt.Printf("Workers conectados: %d/%d\r", current, expectedWorkers)

		if current >= expectedWorkers {
			fmt.Println("\nTodos os workers conectados.")
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Snapshot imutável da lista de workers para evitar race condition
	mu.Lock()
	snapshot := make([]net.Conn, len(workers))
	copy(snapshot, workers)
	mu.Unlock()

	totalWorkers := len(snapshot)

	fmt.Printf("Iniciando brute force distribuído com %d workers\n", totalWorkers)

	// ===== ENCODERS POR WORKER (reutilizados entre tarefas) =====
	// BUG CORRIGIDO: criar json.NewEncoder a cada task corrompia o stream JSON.

	encoders := make([]*json.Encoder, totalWorkers)
	for i, conn := range snapshot {
		encoders[i] = json.NewEncoder(conn)
	}

	// ===== CANAL DE RESULTADO E PROGRESSO GLOBAL =====

	resultChan := make(chan string, 1)
	var globalProgress atomic.Uint64

	// ===== ESCUTAR WORKERS =====

	var listenerWg sync.WaitGroup

	for _, conn := range snapshot {
		listenerWg.Add(1)

		go func(conn net.Conn) {
			defer listenerWg.Done()

			decoder := json.NewDecoder(conn)
			var lastProgress uint64

			for {
				var msg core.Message

				if err := decoder.Decode(&msg); err != nil {
					// Worker desconectou ou encerrou — saída normal
					return
				}

				switch msg.Type {

				case "progress":
					delta := msg.Progress - lastProgress
					lastProgress = msg.Progress
					globalProgress.Add(delta)

				case "result":
					// Envia resultado sem bloquear (outro worker pode ter chegado primeiro)
					select {
					case resultChan <- msg.Result:
					default:
					}
					return
				}
			}
		}(conn)
	}

	// BUG CORRIGIDO: fechar resultChan quando todos os listeners encerrarem,
	// garantindo que o master não bloqueie para sempre quando a senha não é encontrada.
	go func() {
		listenerWg.Wait()
		close(resultChan)
	}()

	// ===== DISTRIBUIR TAREFAS =====

	fmt.Println("Distribuindo tarefas...")

	taskCount := 0

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

			workerIdx := taskCount % totalWorkers

			if err := encoders[workerIdx].Encode(task); err != nil {
				fmt.Printf("Falha ao enviar tarefa para worker %d: %v\n", workerIdx, err)
				continue
			}

			taskCount++
		}
	}

	fmt.Printf("Distribuídas %d tarefas\n", taskCount)

	start := time.Now()

	// ===== PROGRESSO =====

	doneProgress := make(chan struct{})

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		bar := make([]byte, 30)

		for {
			select {

			case <-doneProgress:
				return

			case <-ticker.C:
				current := globalProgress.Load()
				percent := float64(current) / float64(total) * 100

				elapsed := time.Since(start).Seconds()
				hashRate := 0.0
				if elapsed > 0 {
					hashRate = float64(current) / elapsed
				}

				var eta time.Duration
				if hashRate > 0 {
					remaining := float64(total-current) / hashRate
					eta = time.Duration(remaining) * time.Second
				}

				const barSize = 30
				filled := int((percent / 100) * float64(barSize))

				for i := range bar {
					if i < filled {
						bar[i] = '='
					} else {
						bar[i] = ' '
					}
				}

				fmt.Printf(
					"\r[%s] %.4f%% | %d/%d | %.0f H/s | ETA: %s",
					string(bar), percent, current, total,
					hashRate, eta.Truncate(time.Second),
				)
			}
		}
	}()

	// ===== AGUARDAR RESULTADO =====

	result, found := <-resultChan

	close(doneProgress)

	elapsed := time.Since(start).Truncate(time.Millisecond)

	fmt.Println()

	if found && result != "" {
		fmt.Println("Senha encontrada:", result)
	} else {
		fmt.Println("Senha não encontrada")
	}

	fmt.Println("Tempo total:", elapsed)

	// ===== ENCERRAR WORKERS =====

	for _, conn := range snapshot {
		conn.Close()
	}

	fmt.Println("Todos os workers encerrados.")
}
