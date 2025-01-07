// usecase/utils.go
package usecase

import (
	"crypto/sha256"
	"encoding/hex"
)

// GerarSHA256 recebe uma string e retorna seu hash SHA256 em hexadecimal
func GerarSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
