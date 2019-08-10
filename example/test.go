package main

import (
	"crypto/sha512"
	"encoding/hex"
)

func main() {
	for index := 0; index < 5; index++ {
		key := []byte("c:\\/home/akumzy/hola/fs-watcher-go/example/test.go")
		hashByte := sha512.Sum512_224(key)
		hash := hex.EncodeToString(hashByte[:])
		println(hash)
	}

}
