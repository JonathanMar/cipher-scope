package core

type Task struct {
	Prefix     string
	Remaining  int
	Charset    string
	TargetHash string
	HashType   string
}