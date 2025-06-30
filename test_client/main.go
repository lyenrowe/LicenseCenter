package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lyenrowe/LicenseCenter/pkg/crypto"
	"github.com/lyenrowe/LicenseCenter/pkg/utils"
)

// BindFile 绑定请求文件结构
type BindFile struct {
	Hostname    string    `json:"hostname"`
	MachineID   string    `json:"machine_id"`
	RequestTime time.Time `json:"request_time"`
}

// LicenseFile 授权文件结构
type LicenseFile struct {
	LicenseData LicenseData `json:"license_data"`
	Signature   string      `json:"signature"`
}

// LicenseData 授权数据结构
type LicenseData struct {
	MachineID        string    `json:"machine_id"`
	IssuedAt         time.Time `json:"issued_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	LicenseType      string    `json:"license_type"`
	UnbindPrivateKey string    `json:"unbind_private_key"`
}

// UnbindFile 解绑文件结构
type UnbindFile struct {
	SignedLicense LicenseFile `json:"signed_license"`
	UnbindProof   string      `json:"unbind_proof"`
}

// PublicKeyResponse 公钥响应
type PublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run . <action> [options]")
		fmt.Println("可用操作:")
		fmt.Println("  generate-bind         - 生成明文绑定请求文件")
		fmt.Println("  generate-bind-encrypted [server_url] - 生成加密绑定请求文件")
		fmt.Println("  show-machine          - 显示当前机器信息")
		fmt.Println("  decrypt-license <file> - 解密授权文件")
		fmt.Println("  verify-license <file>  - 验证授权文件")
		fmt.Println("  generate-unbind <license_file> - 生成解绑文件")
		return
	}

	action := os.Args[1]

	switch action {
	case "generate-bind":
		generateBindFile()
	case "generate-bind-encrypted":
		serverURL := "http://localhost:8080"
		if len(os.Args) > 2 {
			serverURL = os.Args[2]
		}
		generateEncryptedBindFile(serverURL)
	case "show-machine":
		showMachineInfo()
	case "decrypt-license":
		if len(os.Args) < 3 {
			fmt.Println("请提供授权文件路径")
			return
		}
		decryptLicenseFile(os.Args[2])
	case "verify-license":
		if len(os.Args) < 3 {
			fmt.Println("请提供授权文件路径")
			return
		}
		verifyLicenseFile(os.Args[2])
	case "generate-unbind":
		if len(os.Args) < 3 {
			fmt.Println("请提供授权文件路径")
			return
		}
		generateUnbindFile(os.Args[2])
	default:
		fmt.Println("未知操作:", action)
		fmt.Println("请使用 'go run . --help' 查看帮助")
	}
}

// generateBindFile 生成明文绑定请求文件
func generateBindFile() {
	// 获取机器ID
	machineID, err := utils.GetMachineID()
	if err != nil {
		fmt.Printf("获取机器ID失败: %v\n", err)
		return
	}

	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("获取主机名失败，使用默认值: %v\n", err)
		hostname = "unknown"
	}

	// 创建绑定文件数据
	bindData := BindFile{
		Hostname:    hostname,
		MachineID:   machineID,
		RequestTime: time.Now().UTC(),
	}

	// 序列化为JSON
	fileData, err := json.MarshalIndent(bindData, "", "  ")
	if err != nil {
		fmt.Printf("序列化数据失败: %v\n", err)
		return
	}

	// 生成文件名
	fileName := fmt.Sprintf("%s.bind", hostname)

	// 写入文件
	err = os.WriteFile(fileName, fileData, 0644)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 明文绑定文件生成成功: %s\n", fileName)
	fmt.Printf("📋 文件内容:\n%s\n", string(fileData))
}

// generateEncryptedBindFile 生成加密绑定请求文件
func generateEncryptedBindFile(serverURL string) {
	fmt.Printf("🔄 正在从服务器获取公钥: %s\n", serverURL)

	// 1. 从服务器获取公钥
	publicKey, err := getServerPublicKey(serverURL)
	if err != nil {
		fmt.Printf("❌ 获取服务器公钥失败: %v\n", err)
		return
	}

	fmt.Println("✅ 成功获取服务器公钥")

	// 2. 获取机器信息
	machineID, err := utils.GetMachineID()
	if err != nil {
		fmt.Printf("❌ 获取机器ID失败: %v\n", err)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("⚠️  获取主机名失败，使用默认值: %v\n", err)
		hostname = "unknown"
	}

	// 3. 创建绑定文件数据
	bindData := BindFile{
		Hostname:    hostname,
		MachineID:   machineID,
		RequestTime: time.Now().UTC(),
	}

	// 4. 序列化为JSON
	jsonData, err := json.Marshal(bindData)
	if err != nil {
		fmt.Printf("❌ 序列化数据失败: %v\n", err)
		return
	}

	// 5. 使用混合加密
	encryptedData, err := crypto.EncryptFileToBase64(publicKey, jsonData)
	if err != nil {
		fmt.Printf("❌ 加密数据失败: %v\n", err)
		return
	}

	// 6. 生成加密文件名
	fileName := fmt.Sprintf("%s.bind", hostname)

	// 7. 写入加密文件
	err = os.WriteFile(fileName, []byte(encryptedData), 0644)
	if err != nil {
		fmt.Printf("❌ 写入文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 加密绑定文件生成成功: %s\n", fileName)
	fmt.Printf("📋 原始数据:\n%s\n", string(jsonData))
	fmt.Printf("🔒 文件已加密，内容为Base64编码的密文\n")
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

// decryptLicenseFile 解密授权文件
func decryptLicenseFile(filePath string) {
	fmt.Printf("🔄 正在解密授权文件: %s\n", filePath)

	// 1. 读取文件
	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ 读取文件失败: %v\n", err)
		return
	}

	// 2. 尝试解析为JSON（检查是否为明文文件）
	var licenseFile LicenseFile
	if err := json.Unmarshal(encryptedData, &licenseFile); err == nil {
		// 这是明文文件
		fmt.Println("ℹ️  检测到明文授权文件")
		displayLicenseInfo(licenseFile)
		return
	}

	// 3. 假设是加密文件，需要私钥解密（这里演示，实际应该由服务端处理）
	fmt.Println("🔒 检测到加密文件，需要私钥解密")
	fmt.Println("⚠️  客户端通常不应该持有解密所有文件的私钥")
	fmt.Println("ℹ️  这是演示功能，实际环境中只有服务端能解密")
}

// verifyLicenseFile 验证授权文件
func verifyLicenseFile(filePath string) {
	fmt.Printf("🔄 正在验证授权文件: %s\n", filePath)

	// 读取文件
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ 读取文件失败: %v\n", err)
		return
	}

	// 尝试解析为JSON
	var licenseFile LicenseFile
	if err := json.Unmarshal(fileData, &licenseFile); err != nil {
		fmt.Printf("❌ 解析授权文件失败（可能是加密文件）: %v\n", err)
		return
	}

	// 显示授权信息
	displayLicenseInfo(licenseFile)

	// 检查是否过期
	if time.Now().After(licenseFile.LicenseData.ExpiresAt) {
		fmt.Printf("❌ 授权已过期\n")
	} else {
		fmt.Printf("✅ 授权有效，到期时间: %s\n", licenseFile.LicenseData.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	// 验证机器ID
	currentMachineID, err := utils.GetMachineID()
	if err != nil {
		fmt.Printf("⚠️  无法获取当前机器ID: %v\n", err)
	} else if licenseFile.LicenseData.MachineID == currentMachineID {
		fmt.Printf("✅ 机器ID匹配\n")
	} else {
		fmt.Printf("❌ 机器ID不匹配\n")
		fmt.Printf("   授权机器ID: %s\n", licenseFile.LicenseData.MachineID)
		fmt.Printf("   当前机器ID: %s\n", currentMachineID)
	}
}

// generateUnbindFile 生成解绑文件
func generateUnbindFile(licenseFilePath string) {
	fmt.Printf("🔄 正在生成解绑文件: %s\n", licenseFilePath)

	// 1. 读取授权文件
	fileData, err := os.ReadFile(licenseFilePath)
	if err != nil {
		fmt.Printf("❌ 读取授权文件失败: %v\n", err)
		return
	}

	// 2. 解析授权文件
	var licenseFile LicenseFile
	if err := json.Unmarshal(fileData, &licenseFile); err != nil {
		fmt.Printf("❌ 解析授权文件失败（可能是加密文件）: %v\n", err)
		return
	}

	// 3. 验证机器ID
	currentMachineID, err := utils.GetMachineID()
	if err != nil {
		fmt.Printf("❌ 获取当前机器ID失败: %v\n", err)
		return
	}

	if licenseFile.LicenseData.MachineID != currentMachineID {
		fmt.Printf("❌ 授权文件不属于当前机器\n")
		return
	}

	// 4. 从授权数据中提取解绑私钥
	unbindPrivateKey, err := crypto.LoadPrivateKeyFromPEM(licenseFile.LicenseData.UnbindPrivateKey)
	if err != nil {
		fmt.Printf("❌ 解析解绑私钥失败: %v\n", err)
		return
	}

	// 5. 对整个授权文件进行签名
	licenseDataBytes, err := json.Marshal(licenseFile)
	if err != nil {
		fmt.Printf("❌ 序列化授权文件失败: %v\n", err)
		return
	}

	unbindProof, err := crypto.SignData(unbindPrivateKey, licenseDataBytes)
	if err != nil {
		fmt.Printf("❌ 生成解绑证明失败: %v\n", err)
		return
	}

	// 6. 创建解绑文件
	unbindFile := UnbindFile{
		SignedLicense: licenseFile,
		UnbindProof:   unbindProof,
	}

	// 7. 序列化解绑文件
	unbindData, err := json.MarshalIndent(unbindFile, "", "  ")
	if err != nil {
		fmt.Printf("❌ 序列化解绑文件失败: %v\n", err)
		return
	}

	// 8. 生成文件名
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	fileName := fmt.Sprintf("%s.unbind", hostname)

	// 9. 写入文件
	err = os.WriteFile(fileName, unbindData, 0644)
	if err != nil {
		fmt.Printf("❌ 写入解绑文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 解绑文件生成成功: %s\n", fileName)
	fmt.Printf("⚠️  本地授权现在应该被标记为无效\n")
}

// displayLicenseInfo 显示授权信息
func displayLicenseInfo(licenseFile LicenseFile) {
	fmt.Println("📋 授权文件信息:")
	fmt.Println("================")
	fmt.Printf("机器ID: %s\n", licenseFile.LicenseData.MachineID)
	fmt.Printf("颁发时间: %s\n", licenseFile.LicenseData.IssuedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("到期时间: %s\n", licenseFile.LicenseData.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("授权类型: %s\n", licenseFile.LicenseData.LicenseType)
	fmt.Printf("签名: %s...\n", licenseFile.Signature[:50])
}

// showMachineInfo 显示机器信息
func showMachineInfo() {
	fmt.Println("🖥️  当前机器信息:")
	fmt.Println("================")

	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("主机名: 获取失败 (%v)\n", err)
	} else {
		fmt.Printf("主机名: %s\n", hostname)
	}

	// 获取机器ID
	machineID, err := utils.GetMachineID()
	if err != nil {
		fmt.Printf("机器ID: 获取失败 (%v)\n", err)
	} else {
		fmt.Printf("机器ID: %s\n", machineID)
		fmt.Printf("机器ID (前12位): %s...\n", machineID[:12])
	}

	// 验证机器ID格式
	if machineID != "" {
		if utils.ValidateMachineID(machineID) {
			fmt.Printf("✅ 机器ID格式有效\n")
		} else {
			fmt.Printf("❌ 机器ID格式无效\n")
		}
	}

	fmt.Printf("时间戳: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
