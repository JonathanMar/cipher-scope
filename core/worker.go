package core

import "auditor/crypto"

type Job struct {
	Word       string
	TargetHash string
	HashType   string
}

func Process(job Job) (bool, string) {
	var hash string

	switch job.HashType {
	case "md5":
		hash = crypto.MD5(job.Word)
	case "sha1":
		hash = crypto.SHA1(job.Word)
	case "sha256":
		hash = crypto.SHA256(job.Word)
	default:
		return false, ""
	}

	if hash == job.TargetHash {
		return true, job.Word
	}

	return false, ""
}