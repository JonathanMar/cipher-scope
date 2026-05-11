package main

import (
	"auditor/attacks"
	"auditor/utils"
	"fmt"
	"strings"
)

func main() {
	var targetHash string
	var hashType string

	fmt.Print("Digite o Hash da senha: ")
	_, err := fmt.Scanln(&targetHash)
	if err != nil {
		fmt.Println("Erro ao ler o hash:", err)
		return
	}

	fmt.Print("Digite o Tipo de Criptografia (md5, sha1, sha256): ")
	_, err = fmt.Scanln(&hashType)
	if err != nil {
		fmt.Println("Erro ao ler o tipo:", err)
		return
	}

	targetHash = strings.TrimSpace(targetHash)
	hashType = strings.ToLower(strings.TrimSpace(hashType))

	fmt.Println("\n--- Dados Recebidos ---")
	fmt.Println("Hash:", targetHash)
	fmt.Println("Tipo:", hashType)

	utils.Info("Iniciando ataque de dicionário...")

	result := attacks.DictionaryAttack("wordlist.txt", targetHash, hashType)

	if result != "" {
		utils.Success("Senha encontrada: " + result)
	} else {
		utils.Error("Senha não encontrada")
	}

	fmt.Println("Finalizado.")
}