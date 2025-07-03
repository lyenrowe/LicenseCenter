package crypto

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
)

// RSAKeyPair RSA密钥对
type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// EncryptedFileData 加密文件数据结构
type EncryptedFileData struct {
	EncryptedAESKey []byte // RSA加密的AES密钥
	EncryptedData   []byte // AES加密的JSON数据
}

// GenerateRSAKeyPair 生成RSA密钥对
func GenerateRSAKeyPair(keySize int) (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("生成RSA私钥失败: %w", err)
	}

	return &RSAKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// PrivateKeyToPEM 将私钥转换为PEM格式
func (kp *RSAKeyPair) PrivateKeyToPEM() (string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(kp.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("序列化私钥失败: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

// PublicKeyToPEM 将公钥转换为PEM格式
func (kp *RSAKeyPair) PublicKeyToPEM() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(kp.PublicKey)
	if err != nil {
		return "", fmt.Errorf("序列化公钥失败: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// LoadPrivateKeyFromPEM 从PEM格式加载私钥
func LoadPrivateKeyFromPEM(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM数据")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA私钥")
	}

	return rsaPrivateKey, nil
}

// LoadPublicKeyFromPEM 从PEM格式加载公钥
func LoadPublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM数据")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA公钥")
	}

	return rsaPublicKey, nil
}

// SignData 使用私钥签名数据
func SignData(privateKey *rsa.PrivateKey, data []byte) (string, error) {
	hash := sha256.Sum256(data)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature 使用公钥验证签名
func VerifySignature(publicKey *rsa.PublicKey, data []byte, signature string) error {
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("解码签名失败: %w", err)
	}

	hash := sha256.Sum256(data)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signatureBytes)
	if err != nil {
		return fmt.Errorf("签名验证失败: %w", err)
	}

	return nil
}

// HybridEncrypt 混合加密：使用RSA加密AES密钥，使用AES加密数据
// 参数：
//   - publicKey: RSA公钥，用于加密AES密钥
//   - data: 需要加密的JSON数据
//
// 返回：
//   - []byte: 加密后的数据（格式：[4字节AES密钥长度][RSA加密的AES密钥][AES加密的数据]）
//   - error: 错误信息
func HybridEncrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	// 1. 生成随机AES密钥（32字节，AES-256）
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("生成AES密钥失败: %w", err)
	}

	// 2. 使用AES-GCM加密数据
	encryptedData, err := aesGCMEncrypt(data, aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES加密数据失败: %w", err)
	}

	// 3. 使用RSA-OAEP加密AES密钥
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA加密AES密钥失败: %w", err)
	}

	// 4. 组合数据：[4字节AES密钥长度][RSA加密的AES密钥][AES加密的数据]
	keyLen := len(encryptedAESKey)
	result := make([]byte, 4+keyLen+len(encryptedData))

	// 写入AES密钥长度（大端序）
	binary.BigEndian.PutUint32(result[0:4], uint32(keyLen))

	// 写入加密的AES密钥
	copy(result[4:4+keyLen], encryptedAESKey)

	// 写入加密的数据
	copy(result[4+keyLen:], encryptedData)

	return result, nil
}

// HybridDecrypt 混合解密：使用RSA解密AES密钥，使用AES解密数据
// 参数：
//   - privateKey: RSA私钥，用于解密AES密钥
//   - encryptedData: 加密的数据
//
// 返回：
//   - []byte: 解密后的JSON数据
//   - error: 错误信息
func HybridDecrypt(privateKey *rsa.PrivateKey, encryptedData []byte) ([]byte, error) {
	if len(encryptedData) < 4 {
		return nil, fmt.Errorf("加密数据格式错误：数据太短")
	}

	// 1. 解析数据格式：读取AES密钥长度
	keyLen := binary.BigEndian.Uint32(encryptedData[0:4])
	if len(encryptedData) < int(4+keyLen) {
		return nil, fmt.Errorf("加密数据格式错误：AES密钥数据不完整")
	}

	// 2. 提取RSA加密的AES密钥和AES加密的数据
	encryptedAESKey := encryptedData[4 : 4+keyLen]
	aesEncryptedData := encryptedData[4+keyLen:]

	// 3. 使用RSA-OAEP解密AES密钥
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedAESKey, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA解密AES密钥失败: %w", err)
	}

	// 4. 使用AES-GCM解密数据
	jsonData, err := aesGCMDecrypt(aesEncryptedData, aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES解密数据失败: %w", err)
	}

	return jsonData, nil
}

// aesGCMEncrypt 使用AES-GCM加密数据
func aesGCMEncrypt(data []byte, key []byte) ([]byte, error) {
	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密数据（nonce会被自动添加到密文前面）
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// aesGCMDecrypt 使用AES-GCM解密数据
func aesGCMDecrypt(encryptedData []byte, key []byte) ([]byte, error) {
	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("加密数据太短")
	}

	// 提取nonce和密文
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("GCM解密失败: %w", err)
	}

	return plaintext, nil
}

// EncryptFileToBase64 将JSON数据混合加密后转换为Base64字符串（用于文件存储）
func EncryptFileToBase64(publicKey *rsa.PublicKey, jsonData []byte) (string, error) {
	encryptedData, err := HybridEncrypt(publicKey, jsonData)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// DecryptFileFromBase64 从Base64字符串解密得到JSON数据
func DecryptFileFromBase64(privateKey *rsa.PrivateKey, base64Data string) ([]byte, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %w", err)
	}

	return HybridDecrypt(privateKey, encryptedData)
}

// AESGCMEncrypt 使用AES-GCM加密数据（导出版本）
func AESGCMEncrypt(data []byte, key []byte) ([]byte, error) {
	return aesGCMEncrypt(data, key)
}

// AESGCMDecrypt 使用AES-GCM解密数据（导出版本）
func AESGCMDecrypt(encryptedData []byte, key []byte) ([]byte, error) {
	return aesGCMDecrypt(encryptedData, key)
}

// HybridEncryptWithClientKey 混合加密：使用客户端固定AES密钥
func HybridEncryptWithClientKey(publicKey *rsa.PublicKey, data []byte, clientAESKey []byte) ([]byte, error) {
	// 1. 使用客户端提供的AES密钥加密数据
	encryptedData, err := aesGCMEncrypt(data, clientAESKey)
	if err != nil {
		return nil, fmt.Errorf("AES加密数据失败: %w", err)
	}

	// 2. 使用RSA-OAEP加密客户端AES密钥
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, clientAESKey, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA加密AES密钥失败: %w", err)
	}

	// 3. 组合数据：[4字节AES密钥长度][RSA加密的AES密钥][AES加密的数据]
	keyLen := len(encryptedAESKey)
	result := make([]byte, 4+keyLen+len(encryptedData))

	// 写入AES密钥长度（大端序）
	binary.BigEndian.PutUint32(result[0:4], uint32(keyLen))

	// 写入加密的AES密钥
	copy(result[4:4+keyLen], encryptedAESKey)

	// 写入加密的数据
	copy(result[4+keyLen:], encryptedData)

	return result, nil
}

// EncryptFileToBase64WithClientKey 使用客户端AES密钥的混合加密并转换为Base64
func EncryptFileToBase64WithClientKey(publicKey *rsa.PublicKey, jsonData []byte, clientAESKey []byte) (string, error) {
	encryptedData, err := HybridEncryptWithClientKey(publicKey, jsonData, clientAESKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// GenerateClientAESKey 基于机器ID生成客户端固定AES密钥
func GenerateClientAESKey(machineID string) []byte {
	data := "LicenseCenter:AES:" + machineID
	hash := sha256.Sum256([]byte(data))
	return hash[:] // 返回32字节作为AES-256密钥
}
