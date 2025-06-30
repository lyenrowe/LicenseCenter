package utils

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// GetMachineID 获取机器唯一标识
func GetMachineID() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsMachineID()
	case "linux":
		return getLinuxMachineID()
	case "darwin":
		return getDarwinMachineID()
	default:
		return getFallbackMachineID()
	}
}

// ValidateMachineID 验证机器ID格式
func ValidateMachineID(machineID string) bool {
	// 支持两种格式：
	// 1. MD5格式：32位十六进制字符串
	// 2. SHA256格式：64位十六进制字符串
	machineID = strings.ToLower(machineID)

	// 检查长度和字符
	if len(machineID) == 32 {
		// MD5格式
		matched, _ := regexp.MatchString("^[a-f0-9]{32}$", machineID)
		return matched
	} else if len(machineID) == 64 {
		// SHA256格式
		matched, _ := regexp.MatchString("^[a-f0-9]{64}$", machineID)
		return matched
	}

	return false
}

// getWindowsMachineID 获取Windows机器ID
func getWindowsMachineID() (string, error) {
	var components []string

	// 1. 获取主板序列号
	if mbSerial := getWindowsWMIValue("Win32_BaseBoard", "SerialNumber"); mbSerial != "" {
		components = append(components, "mb:"+mbSerial)
	}

	// 2. 获取系统UUID
	if sysUUID := getWindowsWMIValue("Win32_ComputerSystemProduct", "UUID"); sysUUID != "" {
		components = append(components, "uuid:"+sysUUID)
	}

	// 3. 获取第一个物理硬盘序列号
	if diskSerial := getWindowsWMIValue("Win32_DiskDrive", "SerialNumber"); diskSerial != "" {
		components = append(components, "disk:"+diskSerial)
	}

	// 4. 获取第一个物理网卡MAC地址
	if mac := getFirstPhysicalMACAddress(); mac != "" {
		components = append(components, "mac:"+mac)
	}

	// 如果获取到足够的硬件信息，则组合生成MD5
	if len(components) >= 2 {
		combined := strings.Join(components, "|")
		return hashString(combined), nil
	}

	// 回退方案
	return getFallbackMachineID()
}

// getLinuxMachineID 获取Linux机器ID
func getLinuxMachineID() (string, error) {
	var components []string

	// 1. 尝试读取 /etc/machine-id (systemd)
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		machineID := strings.TrimSpace(string(data))
		if machineID != "" {
			components = append(components, "machine-id:"+machineID)
		}
	}

	// 2. 获取主板序列号
	if mbSerial := getLinuxDMIValue("board_serial"); mbSerial != "" && mbSerial != "None" {
		components = append(components, "mb:"+mbSerial)
	}

	// 3. 获取系统UUID
	if sysUUID := getLinuxDMIValue("product_uuid"); sysUUID != "" && sysUUID != "None" {
		components = append(components, "uuid:"+sysUUID)
	}

	// 4. 获取第一个硬盘序列号
	if diskSerial := getLinuxDiskSerial(); diskSerial != "" {
		components = append(components, "disk:"+diskSerial)
	}

	// 5. 获取第一个物理网卡MAC地址
	if mac := getFirstPhysicalMACAddress(); mac != "" {
		components = append(components, "mac:"+mac)
	}

	// 如果获取到足够的硬件信息，则组合生成MD5
	if len(components) >= 2 {
		combined := strings.Join(components, "|")
		return hashString(combined), nil
	}

	// 回退到读取 /var/lib/dbus/machine-id
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		machineID := strings.TrimSpace(string(data))
		if machineID != "" {
			return hashString(machineID), nil
		}
	}

	// 最终回退方案
	return getFallbackMachineID()
}

// getDarwinMachineID 获取macOS机器ID
func getDarwinMachineID() (string, error) {
	var components []string

	// 1. 获取硬件UUID
	if uuid := getDarwinSystemProfilerValue("Hardware UUID"); uuid != "" {
		components = append(components, "uuid:"+uuid)
	}

	// 2. 获取系统序列号
	if serial := getDarwinSystemProfilerValue("Serial Number"); serial != "" {
		components = append(components, "serial:"+serial)
	}

	// 3. 获取第一个物理网卡MAC地址
	if mac := getFirstPhysicalMACAddress(); mac != "" {
		components = append(components, "mac:"+mac)
	}

	// 如果获取到足够的硬件信息，则组合生成MD5
	if len(components) >= 2 {
		combined := strings.Join(components, "|")
		return hashString(combined), nil
	}

	// 回退方案
	return getFallbackMachineID()
}

// getFallbackMachineID 获取回退机器ID
func getFallbackMachineID() (string, error) {
	var components []string

	// 主机名
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		components = append(components, "hostname:"+hostname)
	}

	// MAC地址
	if mac := getFirstPhysicalMACAddress(); mac != "" {
		components = append(components, "mac:"+mac)
	}

	// 操作系统
	components = append(components, "os:"+runtime.GOOS)

	combined := strings.Join(components, "|")
	return hashString(combined), nil
}

// getFirstPhysicalMACAddress 获取第一个物理网络接口的MAC地址
func getFirstPhysicalMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// 跳过回环接口和虚拟接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 跳过虚拟网络接口（常见的虚拟接口名称）
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "virtual") ||
			strings.Contains(name, "docker") ||
			strings.Contains(name, "bridge") ||
			strings.Contains(name, "veth") ||
			strings.Contains(name, "tap") ||
			strings.Contains(name, "tun") {
			continue
		}

		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String()
		}
	}
	return ""
}

// Windows辅助函数
func getWindowsWMIValue(class, property string) string {
	cmd := exec.Command("wmic", class, "get", property, "/value")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, property+"=") {
			value := strings.TrimPrefix(line, property+"=")
			value = strings.TrimSpace(value)
			if value != "" && value != "None" {
				return value
			}
		}
	}
	return ""
}

// Linux辅助函数
func getLinuxDMIValue(property string) string {
	filePath := "/sys/class/dmi/id/" + property
	if data, err := os.ReadFile(filePath); err == nil {
		value := strings.TrimSpace(string(data))
		if value != "" && value != "None" && !strings.Contains(value, "Not Available") {
			return value
		}
	}
	return ""
}

func getLinuxDiskSerial() string {
	// 尝试从 /sys/block/ 获取第一个物理硬盘的序列号
	cmd := exec.Command("sh", "-c", "find /sys/block -name 'sd*' -o -name 'nvme*' | head -1 | xargs -I {} cat {}/serial 2>/dev/null")
	if output, err := cmd.Output(); err == nil {
		serial := strings.TrimSpace(string(output))
		if serial != "" && serial != "None" {
			return serial
		}
	}

	// 备用方案：使用lsblk
	cmd = exec.Command("lsblk", "-d", "-n", "-o", "SERIAL")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			serial := strings.TrimSpace(line)
			if serial != "" && serial != "None" {
				return serial
			}
		}
	}

	return ""
}

// macOS辅助函数
func getDarwinSystemProfilerValue(property string) string {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, property+":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				value := strings.TrimSpace(parts[1])
				if value != "" && value != "None" {
					return value
				}
			}
		}
	}
	return ""
}

// hashString 对字符串进行MD5哈希
func hashString(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
