package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lyenrowe/LicenseCenter/pkg/crypto"
	_ "github.com/mattn/go-sqlite3"
)

// BindFile 绑定请求文件结构
type BindFile struct {
	Hostname    string    `json:"hostname"`
	MachineID   string    `json:"machine_id"`
	RequestTime time.Time `json:"request_time"`
}

// LicenseFile 授权文件结构 (用于解析license文件)
type LicenseFile struct {
	LicenseData LicenseData `json:"license_data"`
	Signature   string      `json:"signature"`
}

// LicenseData 授权数据结构
type LicenseData struct {
	LicenseKey       string    `json:"license_key"`
	MachineID        string    `json:"machine_id"`
	Hostname         string    `json:"hostname"`
	IssuedAt         time.Time `json:"issued_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	LicenseType      string    `json:"license_type"`
	UnbindPrivateKey string    `json:"unbind_private_key"`
}

// UnbindFile 解绑文件结构
type UnbindFile struct {
	LicenseKey     string         `json:"license_key"`
	MachineID      string         `json:"machine_id"`
	UnbindMetadata UnbindMetadata `json:"unbind_metadata"`
	UnbindProof    string         `json:"unbind_proof"`
}

// UnbindMetadata 解绑元数据
type UnbindMetadata struct {
	UnbindTime    time.Time `json:"unbind_time"`
	Hostname      string    `json:"hostname"`
	ClientVersion string    `json:"client_version"`
	UnbindReason  string    `json:"unbind_reason"`
}

// PublicKeyResponse 公钥响应结构
type PublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	action := os.Args[1]
	switch action {
	case "generate-bind":
		generateBindFiles()
	case "generate-unbind":
		generateUnbindFile()
	case "generate-aes-key":
		generateAESKey()

	case "help", "--help", "-h":
		showHelp()
	default:
		fmt.Printf("❌ 未知操作: %s\n", action)
		showHelp()
	}
}

func showHelp() {
	fmt.Println("🛠️  授权文件测试生成器")
	fmt.Println()
	fmt.Println("用法: go run cmd/test-file-generator/main.go <action> [options]")
	fmt.Println()
	fmt.Println("可用操作:")
	fmt.Println("  generate-bind [count]        生成测试用的 .bind 文件，可选参数count指定生成数量")
	fmt.Println("  generate-unbind <license> [aes_key|machine_id]  根据license文件生成 .unbind 文件")
	fmt.Println("  generate-aes-key <machine_id>  根据机器ID生成对应的AES密钥")

	fmt.Println("  help                         显示此帮助信息")
	fmt.Println()
	fmt.Println("参数说明:")
	fmt.Println("  count                        生成bind文件的数量，默认为1")
	fmt.Println("  license                      license文件的路径（.license或.license.json）")
	fmt.Println("  aes_key|machine_id           可选，客户端AES密钥（Base64）或机器ID（32位十六进制）")
	fmt.Println("  machine_id                   机器ID，用于生成对应的AES密钥")
	fmt.Println()
	fmt.Println("🔒 加密说明:")
	fmt.Println("  - .bind/.unbind 文件为加密版本（可直接用于API）")
	fmt.Println("  - .bind.json/.unbind.json 文件为明文版本（用于调试）")
	fmt.Println("  - bind文件使用基于机器ID的固定AES密钥进行混合加密")
	fmt.Println("  - license文件使用混合加密：RSA加密AES密钥 + AES加密数据")
	fmt.Println("  - 客户端可以基于机器ID重新生成AES密钥来解密license文件")
	fmt.Println("  - 如果不提供参数，工具会尝试使用服务端私钥解密（仅用于调试）")
	fmt.Println("  - ⚠️  客户端不应该持有服务端私钥！")
	fmt.Println("  - 生成器会自动从 http://localhost:8080 获取服务器公钥进行真实加密")
	fmt.Println("  - 如果服务器未运行，将回退到模拟加密（仅用于格式测试）")
	fmt.Println()
	fmt.Println("📁 输出目录:")
	fmt.Println("  - 所有文件生成在 test_data/ 目录下")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-bind")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-bind 3")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-unbind test_data/TEST-PC-01.license.json")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-aes-key abc123def456")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-unbind test_data/TEST-PC-01.license dGVzdF9hZXNfa2V5XzEyMzQ1Njc4")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-unbind test_data/TEST-PC-01.license 2395fe5447a82f1993e4fa244b5220b9")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-unbind test_data/TEST-PC-01.license")
	fmt.Println("  # 可以提供Base64编码的AES密钥、32位十六进制的机器ID，或不提供参数使用服务端私钥解密")
	fmt.Println("  # 使用服务端私钥需要设置环境变量 LICENSE_SERVER_PRIVATE_KEY")
}

// generateBindFiles 生成绑定文件
func generateBindFiles() {
	fmt.Println("🔄 生成测试用 .bind 文件...")

	// 获取生成数量
	count := 1
	if len(os.Args) > 2 {
		if c, err := strconv.Atoi(os.Args[2]); err == nil && c > 0 {
			count = c
		}
	}

	for i := 0; i < count; i++ {
		// 生成虚构的机器信息
		hostname := generateTestHostname(i)
		machineID := generateTestMachineID(hostname, i)

		// 创建绑定文件数据
		bindData := BindFile{
			Hostname:    hostname,
			MachineID:   machineID,
			RequestTime: time.Now().UTC(),
		}

		// 生成明文文件
		if err := saveBindFile(bindData, false); err != nil {
			fmt.Printf("❌ 生成第 %d 个明文 .bind 文件失败: %v\n", i+1, err)
			continue
		}

		// 生成加密文件
		if err := saveBindFile(bindData, true); err != nil {
			fmt.Printf("❌ 生成第 %d 个加密 .bind 文件失败: %v\n", i+1, err)
			continue
		}

		fmt.Printf("✅ 生成第 %d 个 .bind 文件成功: %s (机器ID: %s)\n", i+1, hostname, machineID)
	}

	fmt.Printf("\n🎉 成功生成 %d 组 .bind 文件\n", count)
}

// generateUnbindFile 根据license文件生成解绑文件
func generateUnbindFile() {
	if len(os.Args) < 3 {
		fmt.Println("❌ 请提供license文件路径")
		fmt.Println("用法: go run cmd/test-file-generator/main.go generate-unbind <license_file_path> [aes_key]")
		fmt.Println("参数说明:")
		fmt.Println("  license_file_path    license文件的路径（.license或.license.json）")
		fmt.Println("  aes_key             可选，客户端生成bind文件时使用的AES密钥（Base64编码）")
		return
	}

	licensePath := os.Args[2]
	var aesKey []byte

	// 检查是否提供了AES密钥或机器ID参数
	if len(os.Args) > 3 {
		keyOrMachineID := os.Args[3]

		// 尝试解析为Base64编码的AES密钥
		if decoded, err := base64.StdEncoding.DecodeString(keyOrMachineID); err == nil && len(decoded) == 32 {
			aesKey = decoded
			fmt.Printf("🔑 使用客户端提供的AES密钥进行解密\n")
		} else if len(keyOrMachineID) == 32 {
			// 假设这是32位十六进制的机器ID
			aesKey = crypto.GenerateClientAESKey(keyOrMachineID)
			fmt.Printf("🔑 根据机器ID生成AES密钥进行解密: %s\n", keyOrMachineID)
		} else {
			fmt.Printf("❌ 参数格式错误。应为Base64编码的AES密钥（44字符）或32位十六进制机器ID\n")
			fmt.Printf("   提供的参数: %s (长度: %d)\n", keyOrMachineID, len(keyOrMachineID))
			return
		}
	}

	fmt.Printf("🔄 根据license文件生成 .unbind 文件: %s\n", licensePath)

	// 读取并解析license文件
	licenseData, err := parseLicenseFile(licensePath, aesKey)
	if err != nil {
		fmt.Printf("❌ 解析license文件失败: %v\n", err)
		return
	}

	// 创建解绑文件
	now := time.Now().UTC()

	// 使用license中的私钥生成解绑证明（传入解绑时间和hostname）
	unbindProof, err := generateUnbindProof(licenseData.LicenseKey, licenseData.MachineID, licenseData.Hostname, licenseData.UnbindPrivateKey, now)
	if err != nil {
		fmt.Printf("❌ 生成解绑证明失败: %v\n", err)
		return
	}

	unbindFile := UnbindFile{
		LicenseKey: licenseData.LicenseKey,
		MachineID:  licenseData.MachineID,
		UnbindMetadata: UnbindMetadata{
			UnbindTime:    now,
			Hostname:      licenseData.Hostname,
			ClientVersion: "1.0.0",
			UnbindReason:  "Manual unbind",
		},
		UnbindProof: unbindProof,
	}

	// 保存文件
	if err := saveUnbindFile(unbindFile, licenseData.Hostname, false); err != nil {
		fmt.Printf("❌ 生成明文 .unbind 文件失败: %v\n", err)
		return
	}

	if err := saveUnbindFile(unbindFile, licenseData.Hostname, true); err != nil {
		fmt.Printf("❌ 生成加密 .unbind 文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 生成 .unbind 文件成功: %s (授权码: %s)\n", licenseData.Hostname, licenseData.LicenseKey)
}

// generateAESKey 根据机器ID生成AES密钥
func generateAESKey() {
	if len(os.Args) < 3 {
		fmt.Println("❌ 请提供机器ID")
		fmt.Println("用法: go run cmd/test-file-generator/main.go generate-aes-key <machine_id>")
		fmt.Println("参数说明:")
		fmt.Println("  machine_id    机器ID，用于生成对应的AES密钥")
		return
	}

	machineID := os.Args[2]
	fmt.Printf("🔑 为机器ID生成AES密钥: %s\n", machineID)

	// 使用项目中的算法生成AES密钥
	aesKey := crypto.GenerateClientAESKey(machineID)

	// 转换为Base64编码
	aesKeyBase64 := base64.StdEncoding.EncodeToString(aesKey)

	fmt.Printf("✅ 生成的AES密钥 (Base64编码): %s\n", aesKeyBase64)
	fmt.Printf("📋 可以使用此密钥解密对应机器的license文件:\n")
	fmt.Printf("   go run cmd/test-file-generator/main.go generate-unbind <license_file> %s\n", aesKeyBase64)
}

// parseLicenseFile 解析license文件
func parseLicenseFile(filePath string, aesKey []byte) (*LicenseData, error) {
	// 读取文件
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	var licenseFile LicenseFile

	// 判断是否为加密文件
	if strings.HasSuffix(filePath, ".license.json") {
		// 明文文件，直接解析JSON
		if err := json.Unmarshal(fileData, &licenseFile); err != nil {
			return nil, fmt.Errorf("解析JSON失败: %v", err)
		}
	} else if strings.HasSuffix(filePath, ".license") {
		// 加密文件，需要解密
		decryptedData, err := decryptLicenseFile(fileData, aesKey)
		if err != nil {
			return nil, fmt.Errorf("解密license文件失败: %v", err)
		}

		if err := json.Unmarshal(decryptedData, &licenseFile); err != nil {
			return nil, fmt.Errorf("解析解密后的JSON失败: %v", err)
		}
	} else {
		return nil, fmt.Errorf("不支持的文件格式，请使用 .license 或 .license.json 文件")
	}

	return &licenseFile.LicenseData, nil
}

// decryptLicenseFile 解密license文件
func decryptLicenseFile(encryptedData []byte, aesKey []byte) ([]byte, error) {
	// 如果没有提供客户端AES密钥，尝试通过API解密
	if aesKey == nil {
		return decryptLicenseFileViaAPI(encryptedData)
	}

	// 解析混合加密格式的license文件
	// 格式：[4字节AES密钥长度][RSA加密的AES密钥][AES-GCM加密的JSON数据]
	base64Data := string(encryptedData)
	encryptedBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %v", err)
	}

	// 检查数据长度
	if len(encryptedBytes) < 4 {
		return nil, fmt.Errorf("加密数据格式错误：数据太短")
	}

	// 解析数据格式：读取AES密钥长度
	keyLen := binary.BigEndian.Uint32(encryptedBytes[0:4])
	if len(encryptedBytes) < int(4+keyLen) {
		return nil, fmt.Errorf("加密数据格式错误：AES密钥数据不完整")
	}

	// 跳过RSA加密的AES密钥部分，直接取AES加密的数据
	// 客户端不需要解密RSA部分，因为它可以自己生成AES密钥
	aesEncryptedData := encryptedBytes[4+keyLen:]

	// 使用客户端AES密钥解密数据
	decryptedData, err := crypto.AESGCMDecrypt(aesEncryptedData, aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES解密失败: %v", err)
	}

	return decryptedData, nil
}

// decryptLicenseFileViaAPI 通过API解密license文件
func decryptLicenseFileViaAPI(encryptedData []byte) ([]byte, error) {
	// 尝试使用服务端私钥解密（仅用于调试和测试）
	fmt.Println("⚠️  尝试使用服务端私钥解密（调试模式）")
	fmt.Println("⚠️  注意：客户端不应该持有服务端私钥！")

	decryptedData, err := decryptLicenseFileWithServerKey(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("服务端私钥解密失败: %v", err)
	}

	return decryptedData, nil
}

// decryptLicenseFileWithServerKey 使用服务端私钥解密license文件（仅用于调试）
// 警告：这个方法仅用于开发和调试，客户端不应该持有服务端私钥
func decryptLicenseFileWithServerKey(encryptedData []byte) ([]byte, error) {
	// 解析混合加密格式
	base64Data := string(encryptedData)
	encryptedBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %v", err)
	}

	// 检查数据长度
	if len(encryptedBytes) < 4 {
		return nil, fmt.Errorf("加密数据格式错误：数据太短")
	}

	// 解析数据格式：读取AES密钥长度
	keyLen := binary.BigEndian.Uint32(encryptedBytes[0:4])
	if len(encryptedBytes) < int(4+keyLen) {
		return nil, fmt.Errorf("加密数据格式错误：AES密钥数据不完整")
	}

	// 提取RSA加密的AES密钥和AES加密的数据
	encryptedAESKey := encryptedBytes[4 : 4+keyLen]
	aesEncryptedData := encryptedBytes[4+keyLen:]

	// 尝试获取服务端私钥
	serverPrivateKey, err := getServerPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("获取服务端私钥失败: %v", err)
	}

	// 使用RSA解密AES密钥
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, serverPrivateKey, encryptedAESKey, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA解密AES密钥失败: %v", err)
	}

	// 使用AES解密数据
	decryptedData, err := crypto.AESGCMDecrypt(aesEncryptedData, aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES解密数据失败: %v", err)
	}

	fmt.Printf("🔓 成功使用服务端私钥解密，提取的AES密钥: %s\n", base64.StdEncoding.EncodeToString(aesKey))

	return decryptedData, nil
}

// getServerPrivateKey 获取服务端私钥（仅用于调试）
// 警告：这个方法仅用于开发和调试，客户端不应该持有服务端私钥
func getServerPrivateKey() (*rsa.PrivateKey, error) {
	// 方法1：尝试从本地配置文件读取（如果有的话）
	if privateKey, err := loadPrivateKeyFromConfig(); err == nil {
		return privateKey, nil
	}

	// 方法2：尝试从环境变量读取
	if privateKey, err := loadPrivateKeyFromEnv(); err == nil {
		return privateKey, nil
	}

	// 方法3：尝试从默认数据库路径读取
	if privateKey, err := loadPrivateKeyFromDatabase(); err == nil {
		return privateKey, nil
	}

	return nil, fmt.Errorf("无法获取服务端私钥，请确保：\n" +
		"1. 服务端正在运行，或\n" +
		"2. 设置环境变量 LICENSE_SERVER_PRIVATE_KEY，或\n" +
		"3. 在 data/ 目录下有可访问的数据库文件")
}

// loadPrivateKeyFromConfig 从配置文件加载私钥
func loadPrivateKeyFromConfig() (*rsa.PrivateKey, error) {
	// 这里可以实现从配置文件读取私钥的逻辑
	// 暂时返回错误，表示未实现
	return nil, fmt.Errorf("配置文件方式未实现")
}

// loadPrivateKeyFromEnv 从环境变量加载私钥
func loadPrivateKeyFromEnv() (*rsa.PrivateKey, error) {
	privateKeyPEM := os.Getenv("LICENSE_SERVER_PRIVATE_KEY")
	if privateKeyPEM == "" {
		return nil, fmt.Errorf("环境变量 LICENSE_SERVER_PRIVATE_KEY 未设置")
	}

	privateKey, err := crypto.LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析环境变量中的私钥失败: %v", err)
	}

	return privateKey, nil
}

// loadPrivateKeyFromDatabase 从数据库加载私钥
func loadPrivateKeyFromDatabase() (*rsa.PrivateKey, error) {
	// 尝试从默认数据库路径读取
	dbPath := "data/license.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("数据库文件不存在: %s", dbPath)
	}

	// 连接数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 查询活跃的RSA私钥
	var privateKeyPEM string
	err = db.QueryRow("SELECT private_key FROM rsa_keys WHERE is_active = 1 ORDER BY created_at DESC LIMIT 1").Scan(&privateKeyPEM)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("数据库中没有找到活跃的RSA私钥")
		}
		return nil, fmt.Errorf("查询私钥失败: %v", err)
	}

	// 解析私钥
	privateKey, err := crypto.LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	fmt.Println("🔓 成功从数据库加载服务端私钥")
	return privateKey, nil
}

// generateUnbindProof 生成解绑证明
func generateUnbindProof(licenseKey, machineID, hostname, privateKeyPEM string, unbindTime time.Time) (string, error) {
	// 解析私钥
	privateKey, err := crypto.LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %v", err)
	}

	// 构造待签名的数据（必须与服务端验证时的格式完全一致）
	// 服务端验证格式：licenseKey:machineID:unbindTime(RFC3339):hostname
	unbindData := fmt.Sprintf("%s:%s:%s:%s",
		licenseKey,
		machineID,
		unbindTime.Format(time.RFC3339),
		hostname)

	// 使用私钥签名
	signature, err := crypto.SignData(privateKey, []byte(unbindData))
	if err != nil {
		return "", fmt.Errorf("签名失败: %v", err)
	}

	return signature, nil
}

// 生成测试用主机名
func generateTestHostname(index int) string {
	hostnames := []string{
		"TEST-PC-01", "DEV-WORKSTATION", "QA-MACHINE", "STAGING-SERVER",
		"DEMO-LAPTOP", "BUILD-AGENT", "TEST-NODE", "DEV-CLIENT",
	}

	if index < len(hostnames) {
		return hostnames[index]
	}

	return fmt.Sprintf("TEST-MACHINE-%02d", index+1)
}

// 生成测试用机器ID (MD5格式: 32位十六进制)
func generateTestMachineID(hostname string, index int) string {
	// 使用主机名、索引和当前时间生成唯一的机器ID
	seed := fmt.Sprintf("%s-%d-%d", hostname, index, time.Now().Unix())
	hash := md5.Sum([]byte(seed))
	return fmt.Sprintf("%x", hash)
}

// 保存bind文件
func saveBindFile(bindData BindFile, encrypted bool) error {
	var fileName string
	var content []byte
	var err error

	// 确保test_data目录存在
	if err := os.MkdirAll("test_data", 0755); err != nil {
		return fmt.Errorf("创建test_data目录失败: %v", err)
	}

	if encrypted {
		fileName = fmt.Sprintf("test_data/%s.bind", bindData.Hostname)

		// 获取服务器公钥
		fmt.Printf("📡 正在获取服务器公钥用于加密 %s...\n", bindData.Hostname)
		publicKey, err := getServerPublicKey("http://localhost:8080")
		if err != nil {
			fmt.Printf("⚠️  无法获取服务器公钥，使用模拟加密: %v\n", err)
			// 回退到模拟加密
			jsonData, err := json.Marshal(bindData)
			if err != nil {
				return fmt.Errorf("序列化bind数据失败: %v", err)
			}
			content = []byte(base64.StdEncoding.EncodeToString(jsonData))
		} else {
			// 使用基于机器ID的客户端AES密钥进行混合加密
			jsonData, err := json.Marshal(bindData)
			if err != nil {
				return fmt.Errorf("序列化bind数据失败: %v", err)
			}

			// 生成客户端AES密钥（基于机器ID）
			clientAESKey := crypto.GenerateClientAESKey(bindData.MachineID)

			// 使用客户端AES密钥进行混合加密
			encryptedContent, err := crypto.EncryptFileToBase64WithClientKey(publicKey, jsonData, clientAESKey)
			if err != nil {
				return fmt.Errorf("加密bind数据失败: %v", err)
			}
			content = []byte(encryptedContent)
			fmt.Printf("🔒 使用客户端AES密钥加密生成: %s (机器ID: %s)\n", bindData.Hostname, bindData.MachineID)
		}
	} else {
		fileName = fmt.Sprintf("test_data/%s.bind.json", bindData.Hostname)
		content, err = json.MarshalIndent(bindData, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化bind数据失败: %v", err)
		}
	}

	return os.WriteFile(fileName, content, 0644)
}

// 保存unbind文件
func saveUnbindFile(unbindFile UnbindFile, hostname string, encrypted bool) error {
	var fileName string
	var content []byte
	var err error

	// 确保test_data目录存在
	if err := os.MkdirAll("test_data", 0755); err != nil {
		return fmt.Errorf("创建test_data目录失败: %v", err)
	}

	if encrypted {
		fileName = fmt.Sprintf("test_data/%s.unbind", hostname)

		// 获取服务器公钥
		fmt.Printf("📡 正在获取服务器公钥用于加密 %s.unbind...\n", hostname)
		publicKey, err := getServerPublicKey("http://localhost:8080")
		if err != nil {
			fmt.Printf("⚠️  无法获取服务器公钥，使用模拟加密: %v\n", err)
			// 回退到模拟加密
			jsonData, err := json.Marshal(unbindFile)
			if err != nil {
				return fmt.Errorf("序列化unbind数据失败: %v", err)
			}
			content = []byte(base64.StdEncoding.EncodeToString(jsonData))
		} else {
			// 使用真正的混合加密
			jsonData, err := json.Marshal(unbindFile)
			if err != nil {
				return fmt.Errorf("序列化unbind数据失败: %v", err)
			}

			encryptedContent, err := crypto.EncryptFileToBase64(publicKey, jsonData)
			if err != nil {
				return fmt.Errorf("加密unbind数据失败: %v", err)
			}
			content = []byte(encryptedContent)
			fmt.Printf("🔒 使用真实加密生成: %s.unbind\n", hostname)
		}
	} else {
		fileName = fmt.Sprintf("test_data/%s.unbind.json", hostname)
		content, err = json.MarshalIndent(unbindFile, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化unbind数据失败: %v", err)
		}
	}

	return os.WriteFile(fileName, content, 0644)
}

// getServerPublicKey 从服务器获取公钥
func getServerPublicKey(serverURL string) (*rsa.PublicKey, error) {
	// 构建API URL
	apiURL := strings.TrimSuffix(serverURL, "/") + "/api/public-key"

	// 发送HTTP请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析JSON响应
	var response PublicKeyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 解析公钥
	publicKey, err := crypto.LoadPublicKeyFromPEM(response.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	return publicKey, nil
}
