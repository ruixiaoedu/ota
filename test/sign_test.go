package test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSign(t *testing.T) {

	data := []byte("Hello World")

	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)

	fmt.Println(hex.EncodeToString(hashed))

	hashed2 := sha256.Sum256(data)
	fmt.Println(hex.EncodeToString(hashed2[:]))

	fmt.Printf("%x\n", hashed)
}
