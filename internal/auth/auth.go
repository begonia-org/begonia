package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	mrand "math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spark-lence/tiga"
)

type AuthVildator struct {
	rsaKey *rsa.PrivateKey
}
type AuthSeed struct {
	Seed int64  `json:"seed"`
	Key  string `json:"key"`
}
type AuthKeys struct {
	Pri  string
	Pub  string
	Seed string
}

func NewAuthVildator() *AuthVildator {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	dir := path.Dir(path.Dir(path.Dir(filename)))
	rsaKeyPath := filepath.Join(dir, "cert", "auth_private_key.pem")
	validator := &AuthVildator{}
	rsa, err := validator.LoadPrivateKey(rsaKeyPath)
	if err != nil {
		panic(err)
	}
	validator.rsaKey = rsa
	return validator
}

func (a AuthVildator) GenerateRandPasswd() (string, error) {
	var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var length = 16 // 可以根据需要设置密码长度
	passwd := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("生成随机密码失败,%w", err)
		}
		passwd[i] = charset[num.Int64()]
	}

	return string(passwd), nil
}

func (a AuthVildator) GenerateRandSeed() int64 {
	// 使用当前时间的 Unix 时间戳（秒级）
	timestamp := time.Now().Unix()
	seed := time.Now().UnixNano()

	src := mrand.NewSource(seed)
	r := mrand.New(src)
	// 生成一个 0 到 999 之间的随机数
	randomThreeDigits := r.Intn(10000)

	// 将随机数拼接到时间戳后面
	seed = timestamp*10000 + int64(randomThreeDigits)

	return seed
}

// 生成密钥对
func (a AuthVildator) GenerateKeys(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

// 基于公钥加密
func (a AuthVildator) EncryptWithPublicKey(msg string, pubKey *rsa.PublicKey) ([]byte, error) {
	label := []byte("") // 使用空标签
	hash := sha256.New()

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pubKey, []byte(msg), label)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// 基于私钥解密
func (a AuthVildator) DecryptWithPrivateKey(ciphertext []byte, privKey *rsa.PrivateKey) (string, error) {
	label := []byte("") // 使用空标签
	hash := sha256.New()

	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privKey, ciphertext, label)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
func (a AuthVildator) publicKeyToString(pubKey *rsa.PublicKey) (string, error) {
	// 将公钥转换为 ASN.1 DER 编码
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}

	// 构造 PEM 编码的结构
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY", // 这是一个公钥
		Bytes: derBytes,
	}

	// 将 PEM 编码的数据转换为字符串
	pemBytes := pem.EncodeToMemory(pemBlock)

	return string(pemBytes), nil
}

// 另一种方式是将 DER 编码的数据直接转换为 Base64 编码的字符串
func (a AuthVildator) publicKeyToBase64String(pubKey *rsa.PublicKey) (string, error) {
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(derBytes), nil
}
func (a AuthVildator) EncryptAES(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 随机初始化向量
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 加密器
	stream := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plaintext))
	stream.XORKeyStream(cipherText, []byte(plaintext))
	// 返回带有IV的加密文本
	return hex.EncodeToString(iv) + hex.EncodeToString(cipherText), nil
}
func (a AuthVildator) GenerateAuthSeed(aesKey string) (string, error) {
	key, err := a.GenerateRandPasswd()
	if err != nil {
		return "", err
	}
	seed := AuthSeed{
		Seed: a.GenerateRandSeed(),
		Key:  key,
	}
	str, err := tiga.StructToJsonStr(&seed)
	if err != nil {
		return "", fmt.Errorf("struct转换为json string出错:%w", err)
	}
	encodedData := base64.StdEncoding.EncodeToString([]byte(str))
	return a.EncryptAES([]byte(aesKey), encodedData)

}

// loadPrivateKey 从文件中加载 RSA 私钥
func (a AuthVildator) LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("invalid private key data")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key data: %w", err)
	}
	privKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return privKey, nil
}

// rsaDecrypt 使用 RSA 私钥解密数据
func (a AuthVildator) RSADecrypt(ciphertext string) ([]byte, error) {
	decodedData, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, a.rsaKey, decodedData)
}
