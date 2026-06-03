package core_test

import (
	"auditor/core"
	"testing"
)

func TestValidateHash_MD5(t *testing.T) {
	valid := "5d41402abc4b2a76b9719d911017c592"
	if err := core.ValidateHash("md5", valid); err != nil {
		t.Errorf("ValidateHash(md5, valid) retornou erro: %v", err)
	}
}

func TestValidateHash_NTLM(t *testing.T) {
	valid := "8846f7eaee8fb117ad06bdd830b7586c"
	if err := core.ValidateHash("ntlm", valid); err != nil {
		t.Errorf("ValidateHash(ntlm, valid) retornou erro: %v", err)
	}
}

func TestValidateHash_SHA1(t *testing.T) {
	valid := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
	if err := core.ValidateHash("sha1", valid); err != nil {
		t.Errorf("ValidateHash(sha1, valid) retornou erro: %v", err)
	}
}

func TestValidateHash_SHA256(t *testing.T) {
	valid := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if err := core.ValidateHash("sha256", valid); err != nil {
		t.Errorf("ValidateHash(sha256, valid) retornou erro: %v", err)
	}
}

func TestValidateHash_SHA512(t *testing.T) {
	valid := "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
	if err := core.ValidateHash("sha512", valid); err != nil {
		t.Errorf("ValidateHash(sha512, valid) retornou erro: %v", err)
	}
}

func TestValidateHash_InvalidMD5(t *testing.T) {
	if err := core.ValidateHash("md5", "toocurto"); err == nil {
		t.Error("ValidateHash(md5, curto) deveria retornar erro")
	}
}

func TestValidateHash_InvalidSHA512(t *testing.T) {
	if err := core.ValidateHash("sha512", "abcdef"); err == nil {
		t.Error("ValidateHash(sha512, curto) deveria retornar erro")
	}
}

func TestValidateHash_UnsupportedType(t *testing.T) {
	if err := core.ValidateHash("md3", "abc"); err == nil {
		t.Error("ValidateHash(md3, ...) deveria retornar erro de tipo não suportado")
	}
}

func TestValidateHash_NonHexChars(t *testing.T) {
	nonHex := "gggggggggggggggggggggggggggggggg" // 32 chars mas não hex
	if err := core.ValidateHash("md5", nonHex); err == nil {
		t.Error("ValidateHash(md5, não-hex) deveria retornar erro")
	}
}
