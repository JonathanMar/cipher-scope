package server

import (
	"auditor/crypto"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// ─── Encode / Decode ──────────────────────────────────────────────────────────

type encodeRequest struct {
	Value  string `json:"value"`
	Format string `json:"format"` // base64 | url | hex | binary
	Action string `json:"action"` // encode | decode
}

func HandleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req encodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var result string
	var encErr error

	switch req.Format {
	case "base64":
		if req.Action == "encode" {
			result = base64.StdEncoding.EncodeToString([]byte(req.Value))
		} else {
			b, err := base64.StdEncoding.DecodeString(req.Value)
			if err != nil {
				encErr = err
			} else {
				result = string(b)
			}
		}
	case "url":
		if req.Action == "encode" {
			result = url.QueryEscape(req.Value)
		} else {
			dec, err := url.QueryUnescape(req.Value)
			if err != nil {
				encErr = err
			} else {
				result = dec
			}
		}
	case "hex":
		if req.Action == "encode" {
			result = hex.EncodeToString([]byte(req.Value))
		} else {
			b, err := hex.DecodeString(strings.ReplaceAll(req.Value, " ", ""))
			if err != nil {
				encErr = err
			} else {
				result = string(b)
			}
		}
	case "binary":
		if req.Action == "encode" {
			var sb strings.Builder
			for i, ch := range []byte(req.Value) {
				if i > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteString(fmt.Sprintf("%08b", ch))
			}
			result = sb.String()
		} else {
			parts := strings.Fields(req.Value)
			out := make([]byte, 0, len(parts))
			for _, p := range parts {
				var b byte
				for _, c := range p {
					b = b<<1 | byte(c-'0')
				}
				out = append(out, b)
			}
			result = string(out)
		}
	default:
		http.Error(w, "unsupported format", http.StatusBadRequest)
		return
	}

	if encErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": encErr.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": result})
}

// ─── Hash Identifier ──────────────────────────────────────────────────────────

func HandleIdentifyHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	h := strings.TrimSpace(req.Hash)
	candidates := []string{}

	// Verifica se é hex puro
	isHex := true
	for _, c := range h {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			isHex = false
			break
		}
	}

	if isHex {
		switch len(h) {
		case 32:
			candidates = append(candidates, "md5", "ntlm")
		case 40:
			candidates = append(candidates, "sha1")
		case 64:
			candidates = append(candidates, "sha256")
		case 128:
			candidates = append(candidates, "sha512")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"candidates": candidates,
		"length":     len(h),
		"isHex":      isHex,
	})
}

// ─── Hash Generator (extended) ────────────────────────────────────────────────
// Atualizar HandleHash para suportar sha512 e ntlm.

func HandleHashExtended(w http.ResponseWriter, r *http.Request) {
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
	case "sha512":
		hash = crypto.SHA512(req.Password)
	case "ntlm":
		hash = crypto.NTLM(req.Password)
	default:
		http.Error(w, "unsupported hash type", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"hash": hash})
}

// ─── Chrome DB Decrypter ──────────────────────────────────────────────────────

type ChromePassword struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Error    string `json:"error,omitempty"`
}

func HandleChromeDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Multipart: campo "db" (arquivo) + campo "masterPassword" (opcional)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "falha ao ler formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	masterPassword := r.FormValue("masterPassword")
	if masterPassword == "" {
		masterPassword = "peanuts"
	}

	file, _, err := r.FormFile("db")
	if err != nil {
		http.Error(w, "campo 'db' ausente: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Salva temporariamente
	tmp, err := os.CreateTemp("", "cipher-scope-chrome-*.db")
	if err != nil {
		http.Error(w, "erro interno: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, file); err != nil {
		http.Error(w, "erro ao salvar arquivo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmp.Close()

	// Abre o SQLite
	db, err := sql.Open("sqlite", tmp.Name())
	if err != nil {
		http.Error(w, "não foi possível abrir o banco SQLite: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT origin_url, username_value, password_value FROM logins`)
	if err != nil {
		http.Error(w, "tabela 'logins' não encontrada — certifique-se de usar o arquivo 'Login Data' do Chrome: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer rows.Close()

	var results []ChromePassword
	for rows.Next() {
		var siteURL, username string
		var encPwd []byte
		if err := rows.Scan(&siteURL, &username, &encPwd); err != nil {
			continue
		}
		entry := ChromePassword{URL: siteURL, Username: username}
		plain, err := crypto.ChromeDecrypt(encPwd, masterPassword)
		if err != nil {
			entry.Error = err.Error()
		} else {
			entry.Password = plain
		}
		results = append(results, entry)
	}

	if results == nil {
		results = []ChromePassword{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   len(results),
		"entries": results,
	})
}
