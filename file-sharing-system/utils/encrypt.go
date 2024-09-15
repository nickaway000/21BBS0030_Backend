package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"log"
)

func Encrypt(data []byte, passphrase string) ([]byte, error) {
    block, err := aes.NewCipher([]byte(passphrase))
    if err != nil {
        log.Printf("Encryption error: %v", err)
        return nil, err
    }

    ciphertext := make([]byte, aes.BlockSize+len(data))
    iv := ciphertext[:aes.BlockSize]

    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        log.Printf("IV generation error: %v", err)
        return nil, err
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

    return ciphertext, nil
}


func Decrypt(data []byte, passphrase string) ([]byte, error) {
    key := []byte(passphrase)
    if len(key) != 16 && len(key) != 24 && len(key) != 32 {
        log.Printf("Invalid passphrase length: %d", len(key))
        return nil, errors.New("passphrase must be 16, 24, or 32 bytes long")
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    iv := data[:aes.BlockSize]
    ciphertext := data[aes.BlockSize:]
    log.Printf("IV length: %d, Ciphertext length: %d", len(iv), len(ciphertext))

    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(ciphertext, ciphertext)

    return ciphertext, nil
}

