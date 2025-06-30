package utils

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestGetMachineID 测试机器ID生成
func TestGetMachineID(t *testing.T) {
	machineID, err := GetMachineID()
	if err != nil {
		t.Fatalf("获取机器ID失败: %v", err)
	}

	if machineID == "" {
		t.Fatal("机器ID不能为空")
	}

	if !ValidateMachineID(machineID) {
		t.Fatalf("机器ID格式验证失败: %s", machineID)
	}

	t.Logf("机器ID生成成功: %s", machineID)
}

// TestGetMachineIDDebug 调试模式测试机器ID生成，显示详细的中间过程
func TestGetMachineIDDebug(t *testing.T) {
	fmt.Println("=== 机器ID生成调试信息 ===")
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)
	fmt.Println()

	switch runtime.GOOS {
	case "windows":
		testWindowsMachineIDDebug(t)
	case "linux":
		testLinuxMachineIDDebug(t)
	case "darwin":
		testDarwinMachineIDDebug(t)
	default:
		testFallbackMachineIDDebug(t)
	}

	// 最终生成机器ID
	finalMachineID, err := GetMachineID()
	if err != nil {
		t.Fatalf("最终机器ID生成失败: %v", err)
	}

	fmt.Printf("\n=== 最终结果 ===\n")
	fmt.Printf("生成的机器ID: %s\n", finalMachineID)
	fmt.Printf("机器ID长度: %d位\n", len(finalMachineID))

	if ValidateMachineID(finalMachineID) {
		fmt.Printf("✓ 机器ID格式验证通过\n")
	} else {
		fmt.Printf("✗ 机器ID格式验证失败\n")
		t.Fatalf("机器ID格式验证失败: %s", finalMachineID)
	}
}

// testWindowsMachineIDDebug 测试Windows平台机器ID生成的详细过程
func testWindowsMachineIDDebug(t *testing.T) {
	fmt.Println("=== Windows硬件信息获取 ===")

	// 1. 主板序列号
	mbSerial := getWindowsWMIValue("Win32_BaseBoard", "SerialNumber")
	fmt.Printf("主板序列号: %s\n", displayValue(mbSerial))

	// 2. 系统UUID
	sysUUID := getWindowsWMIValue("Win32_ComputerSystemProduct", "UUID")
	fmt.Printf("系统UUID: %s\n", displayValue(sysUUID))

	// 3. 硬盘序列号
	diskSerial := getWindowsWMIValue("Win32_DiskDrive", "SerialNumber")
	fmt.Printf("硬盘序列号: %s\n", displayValue(diskSerial))

	// 4. 物理网卡MAC
	mac := getFirstPhysicalMACAddress()
	fmt.Printf("物理网卡MAC: %s\n", displayValue(mac))

	// 5. 显示组合信息
	var components []string
	if mbSerial != "" {
		components = append(components, "mb:"+mbSerial)
	}
	if sysUUID != "" {
		components = append(components, "uuid:"+sysUUID)
	}
	if diskSerial != "" {
		components = append(components, "disk:"+diskSerial)
	}
	if mac != "" {
		components = append(components, "mac:"+mac)
	}

	fmt.Printf("\n有效组件数量: %d\n", len(components))
	if len(components) > 0 {
		combined := strings.Join(components, "|")
		fmt.Printf("组合字符串: %s\n", combined)
		fmt.Printf("MD5哈希: %s\n", hashString(combined))
	} else {
		fmt.Println("⚠️ 未获取到有效硬件信息，将使用回退方案")
	}
}

// testLinuxMachineIDDebug 测试Linux平台机器ID生成的详细过程
func testLinuxMachineIDDebug(t *testing.T) {
	fmt.Println("=== Linux硬件信息获取 ===")

	// 1. 系统机器ID
	var machineIDFromFile string
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		machineIDFromFile = strings.TrimSpace(string(data))
	}
	fmt.Printf("/etc/machine-id: %s\n", displayValue(machineIDFromFile))

	// 2. 主板序列号
	mbSerial := getLinuxDMIValue("board_serial")
	fmt.Printf("主板序列号: %s\n", displayValue(mbSerial))

	// 3. 系统UUID
	sysUUID := getLinuxDMIValue("product_uuid")
	fmt.Printf("系统UUID: %s\n", displayValue(sysUUID))

	// 4. 硬盘序列号
	diskSerial := getLinuxDiskSerial()
	fmt.Printf("硬盘序列号: %s\n", displayValue(diskSerial))

	// 5. 物理网卡MAC
	mac := getFirstPhysicalMACAddress()
	fmt.Printf("物理网卡MAC: %s\n", displayValue(mac))

	// 6. 显示组合信息
	var components []string
	if machineIDFromFile != "" {
		components = append(components, "machine-id:"+machineIDFromFile)
	}
	if mbSerial != "" && mbSerial != "None" {
		components = append(components, "mb:"+mbSerial)
	}
	if sysUUID != "" && sysUUID != "None" {
		components = append(components, "uuid:"+sysUUID)
	}
	if diskSerial != "" {
		components = append(components, "disk:"+diskSerial)
	}
	if mac != "" {
		components = append(components, "mac:"+mac)
	}

	fmt.Printf("\n有效组件数量: %d\n", len(components))
	if len(components) > 0 {
		combined := strings.Join(components, "|")
		fmt.Printf("组合字符串: %s\n", combined)
		fmt.Printf("MD5哈希: %s\n", hashString(combined))
	} else {
		fmt.Println("⚠️ 未获取到有效硬件信息，将尝试读取/var/lib/dbus/machine-id")
		if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
			dbusID := strings.TrimSpace(string(data))
			fmt.Printf("/var/lib/dbus/machine-id: %s\n", displayValue(dbusID))
		}
	}
}

// testDarwinMachineIDDebug 测试macOS平台机器ID生成的详细过程
func testDarwinMachineIDDebug(t *testing.T) {
	fmt.Println("=== macOS硬件信息获取 ===")

	// 1. 硬件UUID
	uuid := getDarwinSystemProfilerValue("Hardware UUID")
	fmt.Printf("硬件UUID: %s\n", displayValue(uuid))

	// 2. 系统序列号
	serial := getDarwinSystemProfilerValue("Serial Number")
	fmt.Printf("系统序列号: %s\n", displayValue(serial))

	// 3. 物理网卡MAC
	mac := getFirstPhysicalMACAddress()
	fmt.Printf("物理网卡MAC: %s\n", displayValue(mac))

	// 4. 显示组合信息
	var components []string
	if uuid != "" {
		components = append(components, "uuid:"+uuid)
	}
	if serial != "" {
		components = append(components, "serial:"+serial)
	}
	if mac != "" {
		components = append(components, "mac:"+mac)
	}

	fmt.Printf("\n有效组件数量: %d\n", len(components))
	if len(components) > 0 {
		combined := strings.Join(components, "|")
		fmt.Printf("组合字符串: %s\n", combined)
		fmt.Printf("MD5哈希: %s\n", hashString(combined))
	} else {
		fmt.Println("⚠️ 未获取到有效硬件信息，将使用回退方案")
	}
}

// testFallbackMachineIDDebug 测试回退方案的详细过程
func testFallbackMachineIDDebug(t *testing.T) {
	fmt.Println("=== 回退方案硬件信息获取 ===")

	// 1. 主机名
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	fmt.Printf("主机名: %s\n", displayValue(hostname))

	// 2. 物理网卡MAC
	mac := getFirstPhysicalMACAddress()
	fmt.Printf("物理网卡MAC: %s\n", displayValue(mac))

	// 3. 操作系统
	fmt.Printf("操作系统: %s\n", runtime.GOOS)

	// 4. 显示组合信息
	var components []string
	if hostname != "" {
		components = append(components, "hostname:"+hostname)
	}
	if mac != "" {
		components = append(components, "mac:"+mac)
	}
	components = append(components, "os:"+runtime.GOOS)

	fmt.Printf("\n组件数量: %d\n", len(components))
	combined := strings.Join(components, "|")
	fmt.Printf("组合字符串: %s\n", combined)
	fmt.Printf("MD5哈希: %s\n", hashString(combined))
}

// TestNetworkInterfaces 测试网络接口信息获取
func TestNetworkInterfaces(t *testing.T) {
	fmt.Println("=== 网络接口详细信息 ===")

	interfaces, err := getAllNetworkInterfaces()
	if err != nil {
		t.Fatalf("获取网络接口失败: %v", err)
	}

	fmt.Printf("发现 %d 个网络接口:\n\n", len(interfaces))

	for i, iface := range interfaces {
		fmt.Printf("%d. %s\n", i+1, iface.Name)
		fmt.Printf("   MAC地址: %s\n", displayValue(iface.MAC))
		fmt.Printf("   状态: %s\n", iface.Status)
		fmt.Printf("   类型: %s\n", iface.Type)
		fmt.Printf("   是否物理接口: %t\n", iface.IsPhysical)
		fmt.Println()
	}

	// 显示选中的物理MAC
	physicalMAC := getFirstPhysicalMACAddress()
	fmt.Printf("选中的物理MAC地址: %s\n", displayValue(physicalMAC))
}

// TestValidateMachineID 测试机器ID格式验证
func TestValidateMachineID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效的32位MD5", "1234567890abcdef1234567890abcdef", true},
		{"有效的64位SHA256", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", true},
		{"无效长度", "12345", false},
		{"包含无效字符", "1234567890abcdef1234567890abcdeg", false},
		{"大写字母", "1234567890ABCDEF1234567890ABCDEF", true}, // 会转换为小写
		{"空字符串", "", false},
		{"31位", "1234567890abcdef1234567890abcde", false},
		{"33位", "1234567890abcdef1234567890abcdef1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMachineID(tt.input)
			if result != tt.expected {
				t.Errorf("ValidateMachineID(%q) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

// 辅助函数：显示值，空值时显示特殊标记
func displayValue(value string) string {
	if value == "" {
		return "<空值>"
	}
	if value == "None" {
		return "<None>"
	}
	return value
}

// NetworkInterface 网络接口信息结构
type NetworkInterface struct {
	Name       string
	MAC        string
	Status     string
	Type       string
	IsPhysical bool
}

// getAllNetworkInterfaces 获取所有网络接口的详细信息
func getAllNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []NetworkInterface

	for _, iface := range interfaces {
		ni := NetworkInterface{
			Name: iface.Name,
			MAC:  iface.HardwareAddr.String(),
		}

		// 判断状态
		if iface.Flags&net.FlagUp != 0 {
			ni.Status = "UP"
		} else {
			ni.Status = "DOWN"
		}

		// 判断类型和是否物理接口
		name := strings.ToLower(iface.Name)
		if iface.Flags&net.FlagLoopback != 0 {
			ni.Type = "LOOPBACK"
			ni.IsPhysical = false
		} else if strings.Contains(name, "virtual") ||
			strings.Contains(name, "docker") ||
			strings.Contains(name, "bridge") ||
			strings.Contains(name, "veth") ||
			strings.Contains(name, "tap") ||
			strings.Contains(name, "tun") {
			ni.Type = "VIRTUAL"
			ni.IsPhysical = false
		} else {
			ni.Type = "PHYSICAL"
			ni.IsPhysical = true
		}

		result = append(result, ni)
	}

	return result, nil
}

// BenchmarkGetMachineID 性能基准测试
func BenchmarkGetMachineID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetMachineID()
		if err != nil {
			b.Fatalf("机器ID生成失败: %v", err)
		}
	}
}
