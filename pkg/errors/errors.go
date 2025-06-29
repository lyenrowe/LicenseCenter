package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用程序错误类型
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Code: %d, Message: %s, Error: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

// HTTPStatus 获取HTTP状态码
func (e *AppError) HTTPStatus() int {
	switch {
	case e.Code >= 40000 && e.Code < 41000:
		return http.StatusBadRequest
	case e.Code >= 41000 && e.Code < 42000:
		return http.StatusUnauthorized
	case e.Code >= 42000 && e.Code < 43000:
		return http.StatusForbidden
	case e.Code >= 43000 && e.Code < 44000:
		return http.StatusNotFound
	case e.Code >= 44000 && e.Code < 45000:
		return http.StatusConflict
	case e.Code >= 50000:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// NewAppError 创建新的应用程序错误
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WrapError 包装错误
func WrapError(err error, code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 预定义的错误
var (
	// 认证相关错误 (40xxx)
	ErrInvalidCredentials = NewAppError(40001, "用户名或密码错误")
	ErrInvalidTOTP        = NewAppError(40002, "双因素认证码错误")
	ErrTOTPNotSetup       = NewAppError(40007, "账户未启用双因子认证，请联系管理员")
	ErrTOTPRequired       = NewAppError(40008, "双因子认证码不能为空")
	ErrTOTPForceEnabled   = NewAppError(40009, "系统已启用强制双因子认证，无法禁用")
	ErrTOTPKeyNotSet      = NewAppError(40010, "TOTP密钥未设置")

	// 授权相关错误 (41xxx)
	ErrAuthCodeNotFound  = NewAppError(43001, "授权码不存在")
	ErrAuthCodeDisabled  = NewAppError(41002, "授权码已被禁用")
	ErrInvalidBindFile   = NewAppError(41003, "无效的绑定文件")
	ErrInvalidUnbindFile = NewAppError(41004, "无效的解绑文件")
	ErrInvalidSignature  = NewAppError(41005, "签名验证失败")
	ErrInsufficientSeats = NewAppError(41006, "可用席位不足")
	ErrDuplicateMachine  = NewAppError(41007, "设备已被激活")
	ErrLicenseNotFound   = NewAppError(41008, "授权记录不存在")

	// 加密相关错误 (50xxx)
	ErrCryptoError = NewAppError(50001, "加密操作失败")
)
