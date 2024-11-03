package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"yams/app"
)

func main() {
	bytes, _ := os.ReadFile("app/fooditems.go")

	keyHuman := os.Args[1]
	key, _ := hex.DecodeString(keyHuman)
	fmt.Println("key:", keyHuman)

	final := encrypt(bytes, key)

	err := os.WriteFile("app/fooditems.bin", final, 0600)
	if err != nil {
		panic(err)
	}
}

func encrypt(data []byte, key []byte) []byte {
	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	return gcm.Seal(nonce, nonce, data, nil)
}

func serialize(fooditems []app.FoodItem) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(fooditems)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}
