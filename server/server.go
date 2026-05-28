package server

import (
	"auditor/attacks"
	"auditor/core"
	"auditor/crypto"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ─── State ───────────────────────────────────────────────────────────────────

type Status string

const (
	StatusIdle     Status = "idle"
	StatusRunning  Status = "running"
	StatusFound    Status = "found"
	StatusNotFound Status = "not_found"
)

type attackState struct {
	mu       sync.Mutex
	status   Status
	cancel   context.CancelFunc
	phase    string // "dictionary" | "bruteforce"
	progress atomic.Uint64
	total    atomic.Uint64
	start    time.Time
	result   string
	done     chan struct{} // closed when attack finishes
}

var state = &attackState{status: StatusIdle}

// ─── Snapshot (thread-safe read) ─────────────────────────────────────────────

type StateSnapshot struct {
	Status   Status  `json:"status"`
	Phase    string  `json:"phase"`
	Progress uint64  `json:"progress"`
	Total    uint64  `json:"total"`
	Percent  float64 `json:"percent"`
	HashRate float64 `json:"hashRate"`
	Elapsed  string  `json:"elapsed"`
	ETA      string  `json:"eta"`
	Result   string  `json:"result"`
}

func snapshot() StateSnapshot {
	state.mu.Lock()
	var elapsed string
	if state.status != StatusIdle && !state.start.IsZero() {
		elapsed = time.Since(state.start).Truncate(time.Millisecond).String()
	} else {
		elapsed = "0s"
	}
	s := StateSnapshot{
		Status:  state.status,
		Phase:   state.phase,
		Result:  state.result,
		Elapsed: elapsed,
	}
	state.mu.Unlock()

	s.Progress = state.progress.Load()
	s.Total = state.total.Load()

	if s.Total > 0 {
		s.Percent = float64(s.Progress) / float64(s.Total) * 100
	}

	state.mu.Lock()
	start := state.start
	state.mu.Unlock()

	elapsedSec := time.Since(start).Seconds()
	if elapsedSec > 0 {
		s.HashRate = float64(s.Progress) / elapsedSec
	}
	if s.HashRate > 0 && s.Total > s.Progress {
		remaining := float64(s.Total-s.Progress) / s.HashRate
		s.ETA = time.Duration(remaining * float64(time.Second)).Truncate(time.Second).String()
	}

	return s
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func HandleHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
		Type     string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var hash string
	switch req.Type {
	case "md5":
		hash = crypto.MD5(req.Password)
	case "sha1":
		hash = crypto.SHA1(req.Password)
	case "sha256":
		hash = crypto.SHA256(req.Password)
	default:
		http.Error(w, "unsupported hash type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"hash": hash})
}

func HandleCrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Hash      string `json:"hash"`
		HashType  string `json:"hashType"`
		Mode      string `json:"mode"`
		Workers   int    `json:"workers"`
		MaxLength int    `json:"maxLength"`
		Charset   string `json:"charset"`
		Wordlist  string `json:"wordlist"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := core.ValidateHash(req.HashType, req.Hash); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Workers <= 0 {
		req.Workers = runtime.NumCPU()
	}
	if req.MaxLength <= 0 {
		req.MaxLength = 5
	}
	if req.Charset == "" {
		req.Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	if req.Wordlist == "" {
		req.Wordlist = "wordlist.txt"
	}
	if req.Mode == "" {
		req.Mode = "auto"
	}

	// Only one attack at a time
	state.mu.Lock()
	if state.status == StatusRunning {
		state.mu.Unlock()
		http.Error(w, "attack already running", http.StatusConflict)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	doneCh := make(chan struct{})
	state.status = StatusRunning
	state.cancel = cancel
	state.phase = "dictionary"
	state.result = ""
	state.start = time.Now()
	state.done = doneCh
	state.progress.Store(0)
	state.total.Store(0)
	state.mu.Unlock()

	// Calculate brute-force total for progress reporting
	bfTotal := uint64(0)
	for l := 1; l <= req.MaxLength; l++ {
		bfTotal += uint64(math.Pow(float64(len(req.Charset)), float64(l)))
	}

	onProgress := func() {
		state.progress.Add(1)
	}

	go func() {
		defer cancel()
		defer close(doneCh)

		var result string

		// Phase: dictionary
		if req.Mode == "dictionary" || req.Mode == "auto" {
			state.mu.Lock()
			state.phase = "dictionary"
			state.total.Store(0) // unknown total for dict
			state.mu.Unlock()

			result = attacks.DictionaryAttack(req.Wordlist, req.Hash, req.HashType, onProgress)
		}

		// Phase: brute force
		if result == "" && (req.Mode == "bruteforce" || req.Mode == "auto") {
			state.mu.Lock()
			state.phase = "bruteforce"
			state.progress.Store(0)
			state.total.Store(bfTotal)
			state.start = time.Now() // reset timer for BF phase
			state.mu.Unlock()

			result = attacks.BruteForceAttackUpTo(req.Charset, req.MaxLength, req.Hash, req.HashType, onProgress)
		}

		state.mu.Lock()
		if result != "" {
			state.status = StatusFound
			state.result = result
		} else {
			state.status = StatusNotFound
		}
		state.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func HandleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state.mu.Lock()
	if state.cancel != nil {
		state.cancel()
	}
	state.status = StatusIdle
	state.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	send := func() bool {
		snap := snapshot()
		data, err := json.Marshal(snap)
		if err != nil {
			return true
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		// Stop streaming once a terminal state is reached
		return snap.Status != StatusFound && snap.Status != StatusNotFound
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	ctx := r.Context()

	// Grab the done channel snapshot (may be nil if idle)
	state.mu.Lock()
	doneCh := state.done
	state.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !send() {
				return
			}
		case <-func() <-chan struct{} {
			if doneCh != nil { return doneCh }
			return make(chan struct{}) // never fires
		}():
			send() // send final state
			return
		}
	}
}
