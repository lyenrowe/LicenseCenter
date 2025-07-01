package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/database"
	"github.com/lyenrowe/LicenseCenter/internal/services"
	"github.com/lyenrowe/LicenseCenter/pkg/logger"
	"github.com/skip2/go-qrcode"
	"golang.org/x/term"
)

func main() {
	// 初始化配置
	if err := config.LoadConfig("configs/app.yaml"); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 初始化日志
	if err := logger.InitLogger("info", "logs/app.log"); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}

	// 初始化数据库
	if err := database.InitDatabase(&config.AppConfig.Database); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 执行数据库迁移
	if err := database.DB.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建索引
	if err := database.DB.CreateIndexes(); err != nil {
		log.Fatalf("创建索引失败: %v", err)
	}

	// 创建服务实例
	adminService := services.NewAdminService()
	rsaService := services.NewRSAService()

	fmt.Println("开始初始化系统...")
	fmt.Println("====================================")

	// 1. 生成RSA密钥对
	fmt.Println("生成RSA密钥对...")
	_, _, err := rsaService.GenerateAndSaveKeyPair()
	if err != nil {
		log.Fatalf("生成RSA密钥对失败: %v", err)
	}
	fmt.Println("✓ RSA密钥对生成成功")

	// 2. 交互式创建管理员账户
	fmt.Println("\n创建管理员账户...")
	fmt.Println("====================================")

	// 获取用户名
	username := getInput("请输入管理员用户名 (默认: admin): ")
	if username == "" {
		username = "admin"
	}

	// 获取密码
	password := getPassword("请输入管理员密码 (至少6位): ")
	if len(password) < 6 {
		log.Fatalf("密码长度至少6位")
	}

	// 确认密码
	confirmPassword := getPassword("请再次输入密码确认: ")
	if password != confirmPassword {
		log.Fatalf("两次输入的密码不一致")
	}

	// 创建管理员
	adminReq := &services.CreateAdminRequest{
		Username: username,
		Password: password,
	}

	admin, err := adminService.CreateAdmin(adminReq)
	if err != nil {
		// 如果管理员已存在，跳过
		fmt.Printf("⚠ 管理员创建失败: %v\n", err)
		return
	}

	fmt.Printf("✓ 管理员创建成功 - 用户名: %s\n", admin.Username)

	// 3. 检查是否需要设置双因子认证
	if config.AppConfig.Security.ForceTOTP || admin.TOTPSecret != "" {
		fmt.Println("\n设置双因子认证...")
		fmt.Println("====================================")

		if admin.TOTPSecret != "" {
			// 已经有密钥了，显示二维码
			setupTOTP(adminService, admin.ID)
		} else {
			// 需要生成新的密钥
			fmt.Println("系统启用了强制双因子认证，正在为您设置...")
			setupTOTP(adminService, admin.ID)
		}
	}

	fmt.Println("\n系统初始化完成!")
	fmt.Println("====================================")
	fmt.Printf("管理员账号: %s\n", username)
	if config.AppConfig.Security.ForceTOTP {
		fmt.Println("双因子认证: 已启用")
		fmt.Println("请使用认证器应用扫描上面的二维码完成设置")
	}
	fmt.Println("\n请及时备份您的认证信息!")
}

// getInput 获取用户输入
func getInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// getPassword 安全地获取密码输入（不显示在屏幕上）
func getPassword(prompt string) string {
	fmt.Print(prompt)

	// 使用syscall隐藏密码输入
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("读取密码失败: %v", err)
	}
	fmt.Println() // 添加换行

	return string(bytePassword)
}

// setupTOTP 设置双因子认证
func setupTOTP(adminService *services.AdminService, adminID uint) {
	// 获取TOTP设置信息
	totpInfo, err := adminService.GetTOTPSetupInfo(adminID)
	if err != nil {
		log.Printf("获取TOTP设置信息失败: %v", err)
		return
	}

	var qrCodeURL string

	// 如果还没有TOTP密钥，先生成一个
	if !totpInfo["has_totp_secret"].(bool) {
		url, err := adminService.EnableTOTP(adminID)
		if err != nil {
			log.Printf("启用TOTP失败: %v", err)
			return
		}
		qrCodeURL = url
	} else {
		// 重新获取带二维码的信息
		totpInfo, err = adminService.GetTOTPSetupInfo(adminID)
		if err != nil {
			log.Printf("获取TOTP设置信息失败: %v", err)
			return
		}
		qrCodeURL = totpInfo["qr_code_url"].(string)
	}

	fmt.Println("双因子认证设置:")
	fmt.Println("请使用Google Authenticator、Microsoft Authenticator等应用扫描以下二维码:")
	fmt.Println("")

	// 显示二维码 - 这里我们显示URL，用户可以手动输入到认证器
	fmt.Printf("认证器密钥URL: %s\n", qrCodeURL)
	fmt.Println("")

	// 生成ASCII二维码（如果可能的话）
	displayQRCode(qrCodeURL)

	fmt.Println("设置完成后，请使用认证器生成的6位数字码进行验证:")

	// 验证TOTP设置
	for {
		totpCode := getInput("请输入认证器显示的6位数字码: ")
		if len(totpCode) != 6 {
			fmt.Println("请输入6位数字码")
			continue
		}

		err := adminService.VerifyTOTPSetup(adminID, totpCode)
		if err != nil {
			fmt.Printf("验证失败: %v，请重新输入\n", err)
			continue
		}

		fmt.Println("✓ 双因子认证设置成功!")
		break
	}
}

// displayQRCode 显示二维码
func displayQRCode(url string) {
	fmt.Println("二维码:")
	fmt.Println("========================================")

	// 尝试生成ASCII二维码
	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		fmt.Printf("生成二维码失败: %v\n", err)
		fmt.Println("请手动将以下URL输入到认证器应用中:")
		fmt.Printf("URL: %s\n", url)
		return
	}

	// 生成ASCII版本的二维码
	ascii := q.ToSmallString(false)
	fmt.Println(ascii)

	fmt.Println("========================================")
	fmt.Println("或者您可以使用以下方式:")
	fmt.Println("1. 扫描上面的二维码")
	fmt.Println("2. 手动输入URL到认证器应用")
	fmt.Printf("   URL: %s\n", url)
	fmt.Println("3. 如需更清晰的二维码，可安装qrencode:")
	fmt.Println("   macOS: brew install qrencode")
	fmt.Println("   然后运行: echo '" + url + "' | qrencode -t ANSI")
	fmt.Println("")
}
