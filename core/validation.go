package core

import (
	"errors"
	"regexp"
)

var (
	md5Regex    = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	sha1Regex   = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
	sha256Regex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
)

func ValidateHash(hashType, hash string) error {

	switch hashType {

	case "md5":
		if !md5Regex.MatchString(hash) {
			return errors.New("invalid md5 hash")
		}

	case "sha1":
		if !sha1Regex.MatchString(hash) {
			return errors.New("invalid sha1 hash")
		}

	case "sha256":
		if !sha256Regex.MatchString(hash) {
			return errors.New("invalid sha256 hash")
		}

	default:
		return errors.New("unsupported hash type")
	}

	return nil
}
