package tool

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
)

const (
	PublicKeyStr = `
-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAIcLeIt2wmIyXckgNhCGpMTAZyBGO+nk0/IdOrhIdfRR
gBLHdydsftMVPNHrRuPKQNZRslWE1vvgx80w9lCllIUCAwEAAQ==
-----END PUBLIC KEY-----`
)

// DecryptAes128Ecb 解密阿里云私有加密ts
func DecryptAes128Ecb(data, key []byte) ([]byte, error) {
	decrypt, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decryptData := make([]byte, len(data))
	size := 16
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		decrypt.Decrypt(decryptData[bs:be], data[bs:be])
	}
	return decryptData, nil
}

// DecryptKey 解密视频的AES key
// @param r1    初始随机数
// @param rand  服务端返回的rand
// @param plain 服务端返回的plainText
func DecryptKey(r1, rand, plain string) string {
	r1MD5 := fmt.Sprintf("%x", md5.Sum([]byte(r1)))
	tempKey := r1MD5[8:24]
	iv := []byte(tempKey)
	randDecrypted, _ := Decrypt(iv, iv, rand)
	r2 := r1 + randDecrypted
	r2MD5 := fmt.Sprintf("%x", md5.Sum([]byte(r2)))
	tempKey2 := r2MD5[8:24]
	key2 := []byte(tempKey2)
	finalKey, _ := Decrypt(key2, iv, plain)
	b, _ := base64.StdEncoding.DecodeString(finalKey)
	return fmt.Sprintf("%x", b)
}

// EncryptRand Rand 参数加密
func EncryptRand(origData []byte) (string, error) {
	block, _ := pem.Decode([]byte(PublicKeyStr)) //将密钥解析成公钥实例
	if block == nil {
		return "", errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes) //解析pem.Decode（）返回的Block指针实例
	if err != nil {
		return "", err
	}
	pub := pubInterface.(*rsa.PublicKey)
	data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData) //RSA算法加密
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func Encrypt(key, iv, text []byte) (string, error) {
	//生成cipher.Block 数据块
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("错误 -" + err.Error())
		return "", err
	}
	//填充内容，如果不足16位字符
	blockSize := block.BlockSize()
	originData := pad(text, blockSize)
	//加密方式
	blockMode := cipher.NewCBCEncrypter(block, iv)
	//加密，输出到[]byte数组
	crypted := make([]byte, len(originData))
	blockMode.CryptBlocks(crypted, originData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func pad(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func Decrypt(key, iv []byte, text string) (string, error) {
	decodeData, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", nil
	}
	//生成密码数据块cipher.Block
	block, _ := aes.NewCipher(key)
	//解密模式
	blockMode := cipher.NewCBCDecrypter(block, iv)
	//输出到[]byte数组
	originData := make([]byte, len(decodeData))
	blockMode.CryptBlocks(originData, decodeData)
	//去除填充,并返回
	return string(unPad(originData)), nil
}

func unPad(ciphertext []byte) []byte {
	length := len(ciphertext)
	//去掉最后一次的padding
	unPadding := int(ciphertext[length-1])
	return ciphertext[:(length - unPadding)]
}
