package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	nRead, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("bytes %v", err)
	}
	if nRead < n {
		return nil, fmt.Errorf("Read not enough bits %v", err)
	}
	return b, nil
}

func String(n int) (string, error) {
	b, err := bytes(n)
	if err != nil {
		return "", fmt.Errorf("String: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
