package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
)

// ChromeLinuxKey deriva a chave AES usada pelo Chrome no Linux.
// PBKDF2-HMAC-SHA1(password, "saltysalt", iter=1, keyLen=16).
// Com iter=1: key = HMAC-SHA1(password, salt + \x00\x00\x00\x01)[:16]
func ChromeLinuxKey(password string) []byte {
	h := hmac.New(sha1.New, []byte(password))
	h.Write([]byte("saltysalt"))
	h.Write([]byte{0, 0, 0, 1})
	return h.Sum(nil)[:16]
}

// ChromeDecrypt decripta um password_value BLOB do Chrome (Linux v10/v11).
// masterPassword padrão: "peanuts" (sem keyring do sistema).
func ChromeDecrypt(encryptedValue []byte, masterPassword string) (string, error) {
	if len(encryptedValue) < 4 {
		// Provavelmente texto puro (Chrome antigo)
		return string(encryptedValue), nil
	}

	prefix := string(encryptedValue[:3])
	if prefix != "v10" && prefix != "v11" {
		// Sem prefixo = texto puro
		return string(encryptedValue), nil
	}

	key := ChromeLinuxKey(masterPassword)
	iv := bytes.Repeat([]byte{' '}, 16) // IV = 16 espaços (padrão Chrome Linux)
	ciphertext := encryptedValue[3:]

	if len(ciphertext) == 0 {
		return "", fmt.Errorf("ciphertext vazio")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("tamanho inválido: %d", len(ciphertext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("padding inválido (chave incorreta?): %w", err)
	}

	return string(unpadded), nil
}

// IsPrintable verifica se todos os bytes são ASCII imprimíveis (32–126).
// Usado para validar se a decriptação produziu texto legível.
func IsPrintable(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("dados vazios")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(data) {
		return nil, fmt.Errorf("padding inválido: %d", pad)
	}
	for i := len(data) - pad; i < len(data); i++ {
		if data[i] != byte(pad) {
			return nil, fmt.Errorf("bytes de padding inconsistentes")
		}
	}
	return data[:len(data)-pad], nil
}
