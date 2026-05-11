package main

import (
	"auditor/attacks"
	"auditor/core"
	"auditor/utils"
	"fmt"
	"strings"
	"time"
)

func main() {

	var targetHash string
	var hashType string

	fmt.Print("Digite o hash: ")
	fmt.Scanln(&targetHash)

	fmt.Print("Digite o tipo (md5, sha1, sha256): ")
	fmt.Scanln(&hashType)

	targetHash = strings.TrimSpace(targetHash)
	hashType = strings.ToLower(strings.TrimSpace(hashType))

	if err := core.ValidateHash(hashType, targetHash); err != nil {
		utils.Error(err.Error())
		return
	}

	start := time.Now()

	utils.Info("Iniciando dictionary attack...")

	result := attacks.DictionaryAttack(
		"wordlist.txt",
		targetHash,
		hashType,
	)

	if result != "" {
		utils.Success("Senha encontrada no dicionário: " + result)

		fmt.Printf(
			"Tempo: %s\n",
			time.Since(start),
		)

		return
	}

	utils.Info("Dictionary falhou. Iniciando brute force...")

	result = attacks.BruteForceAttack(
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		5,
		targetHash,
		hashType,
	)

	if result != "" {
		utils.Success("Senha encontrada no brute force: " + result)
	} else {
		utils.Error("Senha não encontrada")
	}

	fmt.Printf(
		"Tempo total: %s\n",
		time.Since(start),
	)
}
