package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
)

func EncryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 数据块填充
	data = pkcs7Padding(data, aes.BlockSize)

	// 加密模式为AES-ECB
	mode := cipher.NewCBCEncrypter(block, make([]byte, aes.BlockSize))

	// 执行加密
	ciphertext := make([]byte, len(data))
	mode.CryptBlocks(ciphertext, data)

	return ciphertext, nil
}

func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 加密模式为AES-ECB
	mode := cipher.NewCBCDecrypter(block, make([]byte, aes.BlockSize))

	// 执行解密
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 数据块去除填充
	plaintext = pkcs7Unpadding(plaintext)

	return plaintext, nil
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

func DeriveKey(password string) []byte {
	h := md5.Sum([]byte(password))
	return h[:]
}
