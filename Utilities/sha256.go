package Utilities

import (
	"crypto/sha256"
	"fmt"
)

func SHA256string(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}
