package utils

import (
	"bufio"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// FileExist 文件是否存在
func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateFile 创建文件（如果目录不存在，同时创建目录）
func CreateFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}

// ParsePrivateKey 解析私钥
func ParsePrivateKey(key []byte) (*rsa.PrivateKey, error) {
	// 解析PEM文件
	block, _ := pem.Decode(key)
	if block == nil {
		// public key error
		return nil, errors.New("this is not the correct key")
	}

	// 解析公钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// ParsePublicKeyFromFile 从文件中解析公钥
func ParsePublicKeyFromFile(filename string) (*rsa.PublicKey, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParsePublicKey(bs)
}

// ParsePublicKey 解析公钥
func ParsePublicKey(key []byte) (*rsa.PublicKey, error) {
	// 解析PEM文件
	block, _ := pem.Decode(key)
	if block == nil {
		// public key error
		return nil, errors.New("this is not the correct key")
	}

	// 解析公钥
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// public key error
		return nil, err
	}

	if rsaPubKey, ok := pubKey.(*rsa.PublicKey); !ok {
		return nil, errors.New("this is not the correct public key")
	} else {
		return rsaPubKey, nil
	}
}

// SignWithSha256 签名
func SignWithSha256(data []byte, prv *rsa.PrivateKey) (string, error) {
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, prv, crypto.SHA256, hashed[:])
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		return "", err
	}
	return hex.EncodeToString(signature), nil
}

// VerifySignWithSha256 验签
func VerifySignWithSha256(data []byte, sign string, pubKey *rsa.PublicKey) bool {

	desSign, err := hex.DecodeString(sign)
	if err != nil {
		return false
	}

	// 验证签名
	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], desSign)
	if err != nil {
		return false
	}
	return true
}

// Md5FromReader 从Read获得MD5值
func Md5FromReader(reader io.Reader) (string, error) {
	r := bufio.NewReader(reader)
	h := md5.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Sha256FromReader 从Read获得SHA256值
func Sha256FromReader(reader io.Reader) (string, error) {
	r := bufio.NewReader(reader)
	h := sha256.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
