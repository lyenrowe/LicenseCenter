package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lyenrowe/LicenseCenter/pkg/utils"
)

// BindFile 绑定请求文件结构
type BindFile struct {
	Hostname    string    `json:"hostname"`
	MachineID   string    `json:"machine_id"`
	RequestTime time.Time `json:"request_time"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run . <action>")
		fmt.Println("可用操作:")
		fmt.Println("  generate-bind  - 生成绑定请求文件")
		fmt.Println("  show-machine   - 显示当前机器信息")
		return
	}

	action := os.Args[1]

	switch action {
	case "generate-bind":
		generateBindFile()
	case "show-machine":
		showMachineInfo()
	default:
		fmt.Println("未知操作:", action)
		fmt.Println("可用操作: generate-bind, show-machine")
	}
}

// generateBindFile 生成绑定请求文件
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

	fmt.Printf("✅ 绑定文件生成成功: %s\n", fileName)
	fmt.Printf("📋 文件内容:\n%s\n", string(fileData))
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
