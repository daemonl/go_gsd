package torch

import (
	"testing"
)

func TestEndToEnd(t *testing.T) {
	plaintext := "abcd1234"
	hashed := HashPassword(plaintext)
	if len(hashed) < 260 {
		t.Error("Insufficient hash length")
	}
	if !CheckPassword(hashed, plaintext) {
		t.Error("Check Password failed for same plaintext")
	}
}

func TestDifferentHashSamePassword(t *testing.T) {
	plaintext := "abcd1234"
	hashed1 := HashPassword(plaintext)
	hashed2 := HashPassword(plaintext)
	if hashed1 == hashed2 {
		t.Error("Two hashed of the same password are equivalent")
	}
}

func TestVariousStrings(t *testing.T) {
	passwords := []string{
		"",
		"abcd",
		"1234",
		"abcd1234abcd1234abdc1234abcd1234abcd1234abcd1234abdc1234abcd1234abcd1234abcd1234abdc1234abcd1234abcd1234abcd1234abdc1234abcd1234",
		"!@3$%^&*()_+=<>,.:\"\\;''`",
		"∆åß∑¨ˆøπ©ƒ∂√ ƒ ¨ˆπ˜√¨∂ƒˆßø˜¢∞¶•ªº£",
	}

	for _, plaintext := range passwords {
		hashed1 := HashPassword(plaintext)
		hashed2 := HashPassword(plaintext)
		if hashed1 == hashed2 {
			t.Errorf("Two hashed of '%s' are equal", plaintext)
		}
		if len(hashed1) < 260 {
			t.Errorf("Insufficient hash length for '%s'", plaintext)
		}
		if !CheckPassword(hashed1, plaintext) {
			t.Errorf("Check Password failed for '%s'", plaintext)
		}
	}
}
