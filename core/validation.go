package core

import (
	"errors"
	"regexp"
)

var (
	md5Regex    = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	sha1Regex   = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
	sha256Regex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	sha512Regex = regexp.MustCompile(`^[a-fA-F0-9]{128}$`)
)

func ValidateHash(hashType, hash string) error {

	switch hashType {

	case "md5", "ntlm":
		// NTLM (NT hash) tem o mesmo formato que MD5: 32 hex chars
		if !md5Regex.MatchString(hash) {
			return errors.New("invalid " + hashType + " hash (expected 32 hex chars)")
		}

	case "sha1":
		if !sha1Regex.MatchString(hash) {
			return errors.New("invalid sha1 hash")
		}

	case "sha256":
		if !sha256Regex.MatchString(hash) {
			return errors.New("invalid sha256 hash")
		}

	case "sha512":
		if !sha512Regex.MatchString(hash) {
			return errors.New("invalid sha512 hash (expected 128 hex chars)")
		}

	default:
		return errors.New("unsupported hash type: " + hashType)
	}

	return nil
}
