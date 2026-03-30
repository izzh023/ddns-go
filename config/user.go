package config

// User 登录用户
type User struct {
	Username string
	Password string
	// 验证码配置
	EnableCaptcha bool   // 是否启用滑动验证码
	CaptchaGT    string // 极验 GT ID
	CaptchaKey   string // 极验 Key
}
