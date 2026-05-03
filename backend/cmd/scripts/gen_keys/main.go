package main

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
)

func main() {
    hashKey := make([]byte, 32)
    encryptKey := make([]byte, 32)
    jwtSecret := make([]byte, 32)

    rand.Read(hashKey)
    rand.Read(encryptKey)
    rand.Read(jwtSecret)

    fmt.Println("COOKIE_HASH_KEY=" + hex.EncodeToString(hashKey))
    fmt.Println("COOKIE_ENCRYPT_KEY=" + hex.EncodeToString(encryptKey))
    fmt.Println("JWT_SECRET=" + hex.EncodeToString(jwtSecret))
}