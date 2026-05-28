package attacks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// hashToJohnFormat mapeia o tipo de hash para o formato do john.
var hashToJohnFormat = map[string]string{
	"md5":    "raw-md5",
	"sha1":   "raw-sha1",
	"sha256": "raw-sha256",
}

// JohnInfo contém informações sobre a instalação do john.
type JohnInfo struct {
	Available bool
	Version   string
	Path      string
}

// CheckJohn verifica se john the ripper está instalado.
func CheckJohn() JohnInfo {
	path, err := exec.LookPath("john")
	if err != nil {
		return JohnInfo{}
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return JohnInfo{Available: true, Path: path}
	}
	ver := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	return JohnInfo{Available: true, Version: ver, Path: path}
}

// JohnAttack usa john the ripper para quebrar o hash.
// useRules=true ativa as regras internas do john (best64).
func JohnAttack(wordlist, targetHash, hashType string, useRules bool) (string, error) {
	format, ok := hashToJohnFormat[hashType]
	if !ok {
		return "", fmt.Errorf("hash type not supported: %s", hashType)
	}

	// Arquivo temporário para o hash
	tmp, err := os.CreateTemp("", "cs-*.hash")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	fmt.Fprintf(tmp, "target:%s\n", targetHash)
	tmp.Close()

	// Pot file isolado para não misturar com o john.pot do usuário
	potFile := tmp.Name() + ".pot"
	defer os.Remove(potFile)

	args := []string{
		"--wordlist=" + wordlist,
		"--format=" + format,
		"--pot=" + potFile,
		"--session=" + tmp.Name(),
		tmp.Name(),
	}
	if useRules {
		args = append(args, "--rules=best64")
	}

	// Captura stdout do john para parsear a senha encontrada
	var stdout strings.Builder
	cmd := exec.Command("john", args...)
	cmd.Stdout = &stdout
	cmd.Run() // ignora código de saída (john sai !=0 quando não encontra)

	// Parseia stdout: john imprime "senha           (target)"
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "(target)") {
			// Remove o sufixo " (target)" e espaços
			pass := strings.TrimSuffix(line, "(target)")
			return strings.TrimSpace(pass), nil
		}
	}

	// Fallback: usa --show para ler o pot file
	show := exec.Command("john", "--show", "--format="+format, "--pot="+potFile, tmp.Name())
	out, err := show.Output()
	if err != nil {
		return "", nil
	}
	// Formato: "target:password::::::"
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), ":", 3)
		if len(parts) >= 2 && parts[0] == "target" && parts[1] != "" {
			return parts[1], nil
		}
	}

	return "", nil
}
