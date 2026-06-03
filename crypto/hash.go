package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"unicode/utf16"
)

func MD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func SHA1(text string) string {
	hash := sha1.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func SHA256(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

func SHA512(text string) string {
	hash := sha512.Sum512([]byte(text))
	return hex.EncodeToString(hash[:])
}

// NTLM calcula o hash NT (MD4 de UTF-16LE) — padrão Windows/Active Directory.
func NTLM(text string) string {
	// Codifica como UTF-16LE
	utf16Chars := utf16.Encode([]rune(text))
	buf := make([]byte, len(utf16Chars)*2)
	for i, c := range utf16Chars {
		buf[i*2] = byte(c)
		buf[i*2+1] = byte(c >> 8)
	}
	return hex.EncodeToString(md4(buf))
}

// md4 implementa o algoritmo MD4 (RFC 1320) necessário para NTLM.
func md4(msg []byte) []byte {
	// Padding
	origLen := len(msg)
	msg = append(msg, 0x80)
	for len(msg)%64 != 56 {
		msg = append(msg, 0x00)
	}
	bitLen := uint64(origLen) * 8
	for i := 0; i < 8; i++ {
		msg = append(msg, byte(bitLen>>(i*8)))
	}

	// Constantes iniciais
	a0, b0, c0, d0 := uint32(0x67452301), uint32(0xEFCDAB89), uint32(0x98BADCFE), uint32(0x10325476)

	for i := 0; i < len(msg)/64; i++ {
		X := make([]uint32, 16)
		for j := 0; j < 16; j++ {
			X[j] = uint32(msg[i*64+j*4]) | uint32(msg[i*64+j*4+1])<<8 |
				uint32(msg[i*64+j*4+2])<<16 | uint32(msg[i*64+j*4+3])<<24
		}
		A, B, C, D := a0, b0, c0, d0
		rot := func(x, n uint32) uint32 { return (x << n) | (x >> (32 - n)) }
		f := func(x, y, z uint32) uint32 { return (x & y) | (^x & z) }
		g := func(x, y, z uint32) uint32 { return (x & y) | (x & z) | (y & z) }
		h := func(x, y, z uint32) uint32 { return x ^ y ^ z }

		// Round 1
		s1 := []uint32{3, 7, 11, 19}
		for _, k := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15} {
			idx := []uint32{0, 1, 2, 3}[(4-k%4)%4]
			_ = idx
			switch k % 4 {
			case 0:
				A = rot(A+f(B, C, D)+X[k], s1[0])
			case 1:
				D = rot(D+f(A, B, C)+X[k], s1[1])
			case 2:
				C = rot(C+f(D, A, B)+X[k], s1[2])
			case 3:
				B = rot(B+f(C, D, A)+X[k], s1[3])
			}
		}
		// Round 2
		s2 := []uint32{3, 5, 9, 13}
		for idx, k := range []int{0, 4, 8, 12, 1, 5, 9, 13, 2, 6, 10, 14, 3, 7, 11, 15} {
			switch idx % 4 {
			case 0:
				A = rot(A+g(B, C, D)+X[k]+0x5A827999, s2[0])
			case 1:
				D = rot(D+g(A, B, C)+X[k]+0x5A827999, s2[1])
			case 2:
				C = rot(C+g(D, A, B)+X[k]+0x5A827999, s2[2])
			case 3:
				B = rot(B+g(C, D, A)+X[k]+0x5A827999, s2[3])
			}
		}
		// Round 3
		s3 := []uint32{3, 9, 11, 15}
		for idx, k := range []int{0, 8, 4, 12, 2, 10, 6, 14, 1, 9, 5, 13, 3, 11, 7, 15} {
			switch idx % 4 {
			case 0:
				A = rot(A+h(B, C, D)+X[k]+0x6ED9EBA1, s3[0])
			case 1:
				D = rot(D+h(A, B, C)+X[k]+0x6ED9EBA1, s3[1])
			case 2:
				C = rot(C+h(D, A, B)+X[k]+0x6ED9EBA1, s3[2])
			case 3:
				B = rot(B+h(C, D, A)+X[k]+0x6ED9EBA1, s3[3])
			}
		}
		a0 += A; b0 += B; c0 += C; d0 += D
	}

	out := make([]byte, 16)
	for i, v := range []uint32{a0, b0, c0, d0} {
		out[i*4] = byte(v); out[i*4+1] = byte(v >> 8)
		out[i*4+2] = byte(v >> 16); out[i*4+3] = byte(v >> 24)
	}
	return out
}
