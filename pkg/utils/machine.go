package utils

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
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
	// 在Windows上使用机器GUID或CPU序列号等
	// 这里简化处理，实际可以调用Windows API
	hostname, _ := os.Hostname()
	return hashString(hostname + "windows"), nil
}

// getLinuxMachineID 获取Linux机器ID
func getLinuxMachineID() (string, error) {
	// 尝试读取 /etc/machine-id
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		machineID := strings.TrimSpace(string(data))
		if machineID != "" {
			return hashString(machineID), nil
		}
	}

	// 尝试读取 /var/lib/dbus/machine-id
	if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		machineID := strings.TrimSpace(string(data))
		if machineID != "" {
			return hashString(machineID), nil
		}
	}

	// 回退到主机名
	return getFallbackMachineID()
}

// getDarwinMachineID 获取macOS机器ID
func getDarwinMachineID() (string, error) {
	// 在macOS上可以使用硬件UUID
	// 这里简化处理，使用主机名和MAC地址组合
	hostname, _ := os.Hostname()
	mac := getFirstMACAddress()
	return hashString(hostname + mac + "darwin"), nil
}

// getFallbackMachineID 获取回退机器ID
func getFallbackMachineID() (string, error) {
	hostname, _ := os.Hostname()
	mac := getFirstMACAddress()
	return hashString(hostname + mac + runtime.GOOS), nil
}

// getFirstMACAddress 获取第一个网络接口的MAC地址
func getFirstMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			if len(iface.HardwareAddr) > 0 {
				return iface.HardwareAddr.String()
			}
		}
	}
	return ""
}

// hashString 对字符串进行MD5哈希
func hashString(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
