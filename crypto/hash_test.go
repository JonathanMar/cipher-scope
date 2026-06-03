package crypto_test

import (
	"auditor/crypto"
	"testing"
)

// ── Vetores conhecidos (valores de referência) ────────────────────────────────

func TestMD5(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"password", "5f4dcc3b5aa765d61d8327deb882cf99"},
		{"The quick brown fox jumps over the lazy dog", "9e107d9d372bb6826bd81d3542a419d6"},
	}
	for _, c := range cases {
		got := crypto.MD5(c.in)
		if got != c.want {
			t.Errorf("MD5(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSHA1(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"password", "5baa61e4c9b93f3f0682250b6cf8331b7ee68fd8"},
	}
	for _, c := range cases {
		got := crypto.SHA1(c.in)
		if got != c.want {
			t.Errorf("SHA1(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSHA256(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"password", "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"},
	}
	for _, c := range cases {
		got := crypto.SHA256(c.in)
		if got != c.want {
			t.Errorf("SHA256(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSHA512(t *testing.T) {
	cases := []struct{ in, want string }{
		{
			"",
			"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
		},
		{
			"hello",
			"9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
		},
		{
			"password",
			"b109f3bbbc244eb82441917ed06d618b9008dd09b3befd1b5e07394c706a8bb980b1d7785e5976ec049b46df5f1326af5a2ea6d103fd07c95385ffab0cacbc86",
		},
	}
	for _, c := range cases {
		got := crypto.SHA512(c.in)
		if got != c.want {
			t.Errorf("SHA512(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNTLM(t *testing.T) {
	// Vetores NTLM oficiais (NT Hash = MD4 de UTF-16LE)
	cases := []struct{ in, want string }{
		// NTLM("") — vazio
		{"", "31d6cfe0d16ae931b73c59d7e0c089c0"},
		// NTLM("password") — vetor clássico
		{"password", "8846f7eaee8fb117ad06bdd830b7586c"},
		// NTLM("Password1") — gerado pelo código
		{"Password1", "64f12cddaa88057e06a81b54e73b949b"},
	}
	for _, c := range cases {
		got := crypto.NTLM(c.in)
		if got != c.want {
			t.Errorf("NTLM(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNTLMLength(t *testing.T) {
	got := crypto.NTLM("qualquer coisa")
	if len(got) != 32 {
		t.Errorf("NTLM deveria retornar 32 chars hex, obteve %d", len(got))
	}
}

func TestChromeLinuxKey(t *testing.T) {
	// A chave deve ter 16 bytes
	key := crypto.ChromeLinuxKey("peanuts")
	if len(key) != 16 {
		t.Errorf("ChromeLinuxKey deve retornar 16 bytes, obteve %d", len(key))
	}
}

func TestChromeDecryptPlaintext(t *testing.T) {
	// Texto puro (Chrome antigo, sem prefixo v10/v11) deve retornar direto
	plain := []byte("minha_senha")
	got, err := crypto.ChromeDecrypt(plain, "peanuts")
	if err != nil {
		t.Fatalf("ChromeDecrypt(plaintext) inesperado erro: %v", err)
	}
	if got != "minha_senha" {
		t.Errorf("ChromeDecrypt(plaintext) = %q, want %q", got, "minha_senha")
	}
}
