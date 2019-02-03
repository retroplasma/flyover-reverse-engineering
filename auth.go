package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "usage: %s [url] [session_id] [token_1] [token_2]\n", os.Args[0])
		os.Exit(1)
	}
	res, err := authURL(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s", res)
}

func authURL(urlStr string, sid string, tokenP1 string, tokenP2 string) (string, error) {
	urlObj, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return "", err
	}

	tokenP3 := genRandStr(16, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	token := tokenP1 + tokenP2 + tokenP3
	ts := time.Now().Unix() + 4200
	path := urlObj.RequestURI()
	sep := map[bool]string{true: "&", false: "?"}[strings.Contains(urlStr, "?")]

	plaintext := fmt.Sprintf("%s%ssid=%s%d%s", path, sep, sid, ts, tokenP3)
	plaintextBytes := padPkcs7([]byte(plaintext))
	key := hashStr(token)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, make([]byte, aes.BlockSize))
	ciphertext := make([]byte, len(plaintextBytes))
	mode.CryptBlocks(ciphertext, plaintextBytes)

	ciphertextEncoded := base64.StdEncoding.EncodeToString([]byte(ciphertext))
	accessKey := fmt.Sprintf("%d_%s_%s", ts, tokenP3, ciphertextEncoded)
	accessKeyEncoded := url.QueryEscape(accessKey)
	final := fmt.Sprintf(`%s%ssid=%s&accessKey=%s`, urlStr, sep, sid, accessKeyEncoded)

	return final, nil
}

func hashStr(str string) []byte {
	hash := sha256.New()
	io.WriteString(hash, str)
	return hash.Sum(nil)
}

func genRandStr(n int, chars string) string {
	b := make([]byte, n)
	for i := range b {
		val, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(err)
		}
		b[i] = chars[val.Int64()]
	}
	return string(b)
}

func padPkcs7(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}
