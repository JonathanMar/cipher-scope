package main

import (
	"auditor/core"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	listenAddr    := flag.String("addr",    ":9000",  "Endereço de escuta (host:porta)")
	hashFlag      := flag.String("hash",    "",       "Hash alvo a quebrar")
	hashTypeFlag  := flag.String("type",    "",       "Tipo de hash: md5, sha1, sha256, sha512, ntlm")
	lengthFlag    := flag.Int("length",     0,        "Tamanho máximo da senha")
	workersFlag   := flag.Int("workers",   1,        "Número de workers a aguardar")
	charsetFlag   := flag.String("charset", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		"Charset para brute force")
	flag.Parse()

	// ===== BANNER =====
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║    Cipher Scope — Master Node        ║")
	fmt.Println("╠══════════════════════════════════════╣")
	fmt.Printf( "║  Escutando em: %-23s║\n", *listenAddr)
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro ao iniciar listener:", err)
		os.Exit(1)
	}
	defer listener.Close()

	var (
		workers []net.Conn
		mu      sync.Mutex
	)

	// ===== ACEITAR WORKERS EM BACKGROUND =====
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			workers = append(workers, conn)
			mu.Unlock()
			fmt.Println("  ✅ Worker conectado:", conn.RemoteAddr())
		}
	}()

	// ===== COLETAR PARÂMETROS (CLI ou interativo) =====

	targetHash     := *hashFlag
	hashType       := *hashTypeFlag
	length         := *lengthFlag
	expectedWorkers := *workersFlag

	if targetHash == "" {
		fmt.Print("Hash alvo: ")
		fmt.Scanln(&targetHash)
	}

	if hashType == "" {
		fmt.Print("Tipo de hash (md5, sha1, sha256, sha512, ntlm): ")
		fmt.Scanln(&hashType)
	}
	hashType = strings.ToLower(strings.TrimSpace(hashType))

	if length == 0 {
		fmt.Print("Tamanho da senha: ")
		fmt.Scanln(&length)
	}

	if *workersFlag == 1 && *hashFlag == "" {
		// Só pergunta se não veio nenhuma flag CLI
		fmt.Print("Quantos workers aguardar? ")
		fmt.Scanln(&expectedWorkers)
	}

	// ===== VALIDAR =====

	if err := core.ValidateHash(hashType, targetHash); err != nil {
		fmt.Fprintln(os.Stderr, "Erro de validação:", err)
		os.Exit(1)
	}

	if length < 2 {
		fmt.Fprintln(os.Stderr, "Tamanho mínimo da senha é 2")
		os.Exit(1)
	}

	charset := *charsetFlag

	// Resumo do ataque
	fmt.Println()
	fmt.Printf("  Hash:    %s\n", targetHash)
	fmt.Printf("  Tipo:    %s\n", hashType)
	fmt.Printf("  Tamanho: %d\n", length)
	fmt.Printf("  Workers: %d\n", expectedWorkers)
	fmt.Printf("  Charset: %s\n", charset)
	fmt.Println()

	total := uint64(math.Pow(float64(len(charset)), float64(length)))

	// ===== AGUARDAR WORKERS =====

	fmt.Printf("Aguardando %d worker(s)...\n", expectedWorkers)

	for {
		mu.Lock()
		current := len(workers)
		mu.Unlock()

		fmt.Printf("  Workers conectados: %d/%d\r", current, expectedWorkers)

		if current >= expectedWorkers {
			fmt.Printf("\n  Todos os workers conectados.\n\n")
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Snapshot imutável para evitar race condition
	mu.Lock()
	snapshot := make([]net.Conn, len(workers))
	copy(snapshot, workers)
	mu.Unlock()

	totalWorkers := len(snapshot)

	// ===== ENCODERS POR WORKER =====
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

	fmt.Printf("Distribuindo tarefas entre %d worker(s)...\n", totalWorkers)

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

	fmt.Printf("  %d tarefas distribuídas\n\n", taskCount)

	start := time.Now()

	// ===== BARRA DE PROGRESSO =====

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
					"\r[%s] %.2f%% | %d/%d | %.0f H/s | ETA: %s",
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
	fmt.Println()

	if found && result != "" {
		fmt.Printf("  ✅ Senha encontrada: %s\n", result)
	} else {
		fmt.Println("  ❌ Senha não encontrada")
	}

	fmt.Printf("  ⏱  Tempo total: %s\n", elapsed)

	// ===== ENCERRAR WORKERS =====

	for _, conn := range snapshot {
		conn.Close()
	}

	fmt.Println("  Todos os workers encerrados.")
}
