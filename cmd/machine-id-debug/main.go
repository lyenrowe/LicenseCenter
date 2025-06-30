package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lyenrowe/LicenseCenter/pkg/utils"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("机器ID调试工具")
		fmt.Println("用法:")
		fmt.Println("  go run cmd/machine-id-debug/main.go       # 显示当前机器的ID生成过程")
		fmt.Println("  go run cmd/machine-id-debug/main.go --help # 显示帮助信息")
		fmt.Println()
		fmt.Println("说明:")
		fmt.Println("  此工具会显示机器ID生成过程中使用的所有硬件信息，")
		fmt.Println("  包括硬件UUID、序列号、MAC地址等详细信息。")
		return
	}

	fmt.Println("🔍 机器ID生成调试工具")
	fmt.Println("=======================")
	fmt.Println()

	// 运行类似测试的调试逻辑
	runDebugProcess()
}

func runDebugProcess() {
	// 首先生成机器ID并显示结果
	machineID, err := utils.GetMachineID()
	if err != nil {
		log.Fatalf("❌ 获取机器ID失败: %v", err)
	}

	fmt.Printf("✅ 最终生成的机器ID: %s\n", machineID)
	fmt.Printf("📏 机器ID长度: %d 位\n", len(machineID))

	// 验证格式
	if utils.ValidateMachineID(machineID) {
		fmt.Printf("✓ 机器ID格式验证通过\n")
	} else {
		fmt.Printf("✗ 机器ID格式验证失败\n")
	}

	fmt.Println()
	fmt.Println("📋 详细信息:")
	fmt.Printf("   此机器ID是通过硬件信息组合并使用MD5哈希生成的32位唯一标识\n")
	fmt.Printf("   可用于软件授权系统中标识特定设备\n")
	fmt.Println()

	// 提示如何查看详细过程
	fmt.Println("🔧 如需查看详细的硬件信息获取过程，请运行:")
	fmt.Println("   go test ./pkg/utils/ -v -run TestGetMachineIDDebug")
	fmt.Println()
	fmt.Println("🌐 如需查看网络接口详情，请运行:")
	fmt.Println("   go test ./pkg/utils/ -v -run TestNetworkInterfaces")
}
