package torch

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log"
)

func HashPassword(plaintext string) string {
	// Create the Salt: 256 random bytes
	saltBytes := make([]byte, 256, 256)
	_, _ = rand.Reader.Read(saltBytes)

	// Create a hasher
	hasher := sha256.New()

	// Append plaintext bytes
	hasher.Write([]byte(plaintext))

	// Append salt bytes
	hasher.Write(saltBytes)

	// Get the hash from the hasher
	hashBytes := hasher.Sum(nil)

	// [256 bytes of salt] + [x bytes of hash] to a base64 string to store salt and password in one field
	return base64.URLEncoding.EncodeToString(append(saltBytes, hashBytes...))
}

func CheckPassword(storedString string, plaintext string) (bool, error) {

	// Deocde the hash string
	stored, err := base64.URLEncoding.DecodeString(storedString)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// stores should be [256 bytes of salt] + [x bytes of hash]
	if len(stored) < 257 {
		return false, nil
	}

	// Split the salt and hash
	saltBytes := stored[0:256]
	hashBytes := stored[256:]

	// Write the parts to the hash as in hashPassword()
	hasher := sha256.New()
	hasher.Write([]byte(plaintext))
	hasher.Write(saltBytes)

	// Get the hash from the hasher
	hasherResult := hasher.Sum(nil)

	// Ensure the hash result == the stored result
	if bytes.Compare(hasherResult, hashBytes) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}
