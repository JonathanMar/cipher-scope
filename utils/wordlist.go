package utils

import (
	"bufio"
	"os"
)

func LoadWordlist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	return words, scanner.Err()
}
