package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lyenrowe/LicenseCenter/pkg/crypto"
)

// BindFile 绑定请求文件结构
type BindFile struct {
	Hostname    string    `json:"hostname"`
	MachineID   string    `json:"machine_id"`
	RequestTime time.Time `json:"request_time"`
}

// LicenseFile 授权文件结构 (用于生成unbind文件)
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
	case "generate-license":
		generateLicenseFile()
	case "generate-unbind":
		generateUnbindFile()
	case "generate-all":
		generateAllFiles()
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
	fmt.Println("  generate-bind     生成测试用的 .bind 文件")
	fmt.Println("  generate-license  生成测试用的 .license 文件")
	fmt.Println("  generate-unbind   生成测试用的 .unbind 文件")
	fmt.Println("  generate-all      生成完整的测试文件集合")
	fmt.Println("  help              显示此帮助信息")
	fmt.Println()
	fmt.Println("🔒 加密说明:")
	fmt.Println("  - .bind/.license/.unbind 文件为加密版本（可直接用于API激活）")
	fmt.Println("  - .bind.json/.license.json/.unbind.json 文件为明文版本（用于调试）")
	fmt.Println("  - 生成器会自动从 http://localhost:8080 获取服务器公钥进行真实加密")
	fmt.Println("  - 如果服务器未运行，将回退到模拟加密（仅用于格式测试）")
	fmt.Println()
	fmt.Println("📁 输出目录:")
	fmt.Println("  - 所有文件生成在 test_data/ 目录下")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-bind")
	fmt.Println("  go run cmd/test-file-generator/main.go generate-all")
	fmt.Println()
	fmt.Println("🧪 测试激活:")
	fmt.Println("  ./scripts/test_activation.sh")
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

// generateLicenseFile 生成测试授权文件
func generateLicenseFile() {
	fmt.Println("🔄 生成测试用 .license 文件...")

	// 生成虚构的机器信息
	hostname := generateTestHostname(0)
	machineID := generateTestMachineID(hostname, 0)

	// 生成解绑密钥对
	unbindKeyPair, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		fmt.Printf("❌ 生成解绑密钥对失败: %v\n", err)
		return
	}

	unbindPrivateKeyPEM, err := unbindKeyPair.PrivateKeyToPEM()
	if err != nil {
		fmt.Printf("❌ 转换解绑私钥失败: %v\n", err)
		return
	}

	// 创建授权数据
	now := time.Now().UTC()
	expiresAt := now.AddDate(1, 0, 0) // 1年后过期
	licenseKey := fmt.Sprintf("LIC-%s-%d", machineID[:8], now.Unix())
	licenseData := LicenseData{
		LicenseKey:       licenseKey,
		MachineID:        machineID,
		Hostname:         hostname,
		IssuedAt:         now,
		ExpiresAt:        expiresAt,
		LicenseType:      "FULL",
		UnbindPrivateKey: unbindPrivateKeyPEM,
	}

	// 创建模拟签名（用于测试）
	signature := generateTestSignature(machineID)

	licenseFile := LicenseFile{
		LicenseData: licenseData,
		Signature:   signature,
	}

	// 保存文件
	if err := saveLicenseFile(licenseFile, hostname, false); err != nil {
		fmt.Printf("❌ 生成明文 .license 文件失败: %v\n", err)
		return
	}

	if err := saveLicenseFile(licenseFile, hostname, true); err != nil {
		fmt.Printf("❌ 生成加密 .license 文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 生成 .license 文件成功: %s (机器ID: %s)\n", hostname, machineID)
	fmt.Printf("📅 授权期限: %s - %s\n", now.Format("2006-01-02"), expiresAt.Format("2006-01-02"))
}

// generateUnbindFile 生成解绑文件
func generateUnbindFile() {
	fmt.Println("🔄 生成测试用 .unbind 文件...")

	// 先生成一个测试license文件
	hostname := generateTestHostname(0)
	machineID := generateTestMachineID(hostname, 0)

	// 生成测试用的license_key
	now := time.Now().UTC()
	licenseKey := fmt.Sprintf("LIC-%s-%d", machineID[:8], now.Unix())

	// 生成解绑证明（模拟签名）
	unbindProof := generateTestSignature(machineID + "-unbind")

	unbindFile := UnbindFile{
		LicenseKey:     licenseKey,
		MachineID:      machineID,
		UnbindMetadata: UnbindMetadata{UnbindTime: now, Hostname: hostname, ClientVersion: "1.0.0", UnbindReason: "Test"},
		UnbindProof:    unbindProof,
	}

	// 保存文件
	if err := saveUnbindFile(unbindFile, hostname, false); err != nil {
		fmt.Printf("❌ 生成明文 .unbind 文件失败: %v\n", err)
		return
	}

	if err := saveUnbindFile(unbindFile, hostname, true); err != nil {
		fmt.Printf("❌ 生成加密 .unbind 文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 生成 .unbind 文件成功: %s (机器ID: %s)\n", hostname, machineID)
}

// generateAllFiles 生成完整的测试文件集合
func generateAllFiles() {
	fmt.Println("🔄 生成完整的测试文件集合...")

	// 生成一组相关的测试文件
	hostname := generateTestHostname(0)
	machineID := generateTestMachineID(hostname, 0)

	fmt.Printf("📋 测试设备信息:\n")
	fmt.Printf("   主机名: %s\n", hostname)
	fmt.Printf("   机器ID: %s\n", machineID)
	fmt.Println()

	// 1. 生成bind文件
	fmt.Println("1️⃣  生成 .bind 文件...")
	bindData := BindFile{
		Hostname:    hostname,
		MachineID:   machineID,
		RequestTime: time.Now().UTC(),
	}

	if err := saveBindFile(bindData, false); err != nil {
		fmt.Printf("❌ 生成明文 .bind 文件失败: %v\n", err)
		return
	}

	if err := saveBindFile(bindData, true); err != nil {
		fmt.Printf("❌ 生成加密 .bind 文件失败: %v\n", err)
		return
	}

	// 2. 生成license文件
	fmt.Println("2️⃣  生成 .license 文件...")
	unbindKeyPair, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		fmt.Printf("❌ 生成解绑密钥对失败: %v\n", err)
		return
	}

	unbindPrivateKeyPEM, err := unbindKeyPair.PrivateKeyToPEM()
	if err != nil {
		fmt.Printf("❌ 转换解绑私钥失败: %v\n", err)
		return
	}

	now := time.Now().UTC()
	licenseKey := fmt.Sprintf("LIC-%s-%d", machineID[:8], now.Unix())
	licenseData := LicenseData{
		LicenseKey:       licenseKey,
		MachineID:        machineID,
		Hostname:         hostname,
		IssuedAt:         now,
		ExpiresAt:        now.AddDate(1, 0, 0),
		LicenseType:      "FULL",
		UnbindPrivateKey: unbindPrivateKeyPEM,
	}

	licenseFile := LicenseFile{
		LicenseData: licenseData,
		Signature:   generateTestSignature(machineID),
	}

	if err := saveLicenseFile(licenseFile, hostname, false); err != nil {
		fmt.Printf("❌ 生成明文 .license 文件失败: %v\n", err)
		return
	}

	if err := saveLicenseFile(licenseFile, hostname, true); err != nil {
		fmt.Printf("❌ 生成加密 .license 文件失败: %v\n", err)
		return
	}

	// 3. 生成unbind文件
	fmt.Println("3️⃣  生成 .unbind 文件...")
	unbindFile := UnbindFile{
		LicenseKey:     licenseKey,
		MachineID:      machineID,
		UnbindMetadata: UnbindMetadata{UnbindTime: now, Hostname: hostname, ClientVersion: "1.0.0", UnbindReason: "Test"},
		UnbindProof:    generateTestSignature(machineID + "-unbind"),
	}

	if err := saveUnbindFile(unbindFile, hostname, false); err != nil {
		fmt.Printf("❌ 生成明文 .unbind 文件失败: %v\n", err)
		return
	}

	if err := saveUnbindFile(unbindFile, hostname, true); err != nil {
		fmt.Printf("❌ 生成加密 .unbind 文件失败: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("🎉 完整测试文件集合生成成功！\n")
	fmt.Printf("📁 生成的文件 (位于 test_data/ 目录):\n")
	fmt.Printf("   - test_data/%s.bind.json (明文)\n", hostname)
	fmt.Printf("   - test_data/%s.bind (加密)\n", hostname)
	fmt.Printf("   - test_data/%s.license.json (明文)\n", hostname)
	fmt.Printf("   - test_data/%s.license (加密)\n", hostname)
	fmt.Printf("   - test_data/%s.unbind.json (明文)\n", hostname)
	fmt.Printf("   - test_data/%s.unbind (加密)\n", hostname)
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

// 生成测试用签名
func generateTestSignature(data string) string {
	// 生成一个随机的Base64字符串作为模拟签名
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	signature := fmt.Sprintf("TEST-SIGNATURE-%s-%d", data[:8], n.Int64())
	return base64.StdEncoding.EncodeToString([]byte(signature))
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
			// 使用真正的混合加密
			jsonData, err := json.Marshal(bindData)
			if err != nil {
				return fmt.Errorf("序列化bind数据失败: %v", err)
			}

			encryptedContent, err := crypto.EncryptFileToBase64(publicKey, jsonData)
			if err != nil {
				return fmt.Errorf("加密bind数据失败: %v", err)
			}
			content = []byte(encryptedContent)
			fmt.Printf("🔒 使用真实加密生成: %s\n", bindData.Hostname)
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

// 保存license文件
func saveLicenseFile(licenseFile LicenseFile, hostname string, encrypted bool) error {
	var fileName string
	var content []byte
	var err error

	// 确保test_data目录存在
	if err := os.MkdirAll("test_data", 0755); err != nil {
		return fmt.Errorf("创建test_data目录失败: %v", err)
	}

	if encrypted {
		fileName = fmt.Sprintf("test_data/%s.license", hostname)

		// 获取服务器公钥
		fmt.Printf("📡 正在获取服务器公钥用于加密 %s.license...\n", hostname)
		publicKey, err := getServerPublicKey("http://localhost:8080")
		if err != nil {
			fmt.Printf("⚠️  无法获取服务器公钥，使用模拟加密: %v\n", err)
			// 回退到模拟加密
			jsonData, err := json.Marshal(licenseFile)
			if err != nil {
				return fmt.Errorf("序列化license数据失败: %v", err)
			}
			content = []byte(base64.StdEncoding.EncodeToString(jsonData))
		} else {
			// 使用真正的混合加密
			jsonData, err := json.Marshal(licenseFile)
			if err != nil {
				return fmt.Errorf("序列化license数据失败: %v", err)
			}

			encryptedContent, err := crypto.EncryptFileToBase64(publicKey, jsonData)
			if err != nil {
				return fmt.Errorf("加密license数据失败: %v", err)
			}
			content = []byte(encryptedContent)
			fmt.Printf("🔒 使用真实加密生成: %s.license\n", hostname)
		}
	} else {
		fileName = fmt.Sprintf("test_data/%s.license.json", hostname)
		content, err = json.MarshalIndent(licenseFile, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化license数据失败: %v", err)
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
