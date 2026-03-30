package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/jeessy2/ddns-go/v6/config"
	"github.com/jeessy2/ddns-go/v6/util"
)

//go:embed login_new.html
var loginNewEmbedFile embed.FS

// LoginConfig 登录配置
type LoginConfig struct {
	EnableCaptcha bool   // 是否启用验证码
	CaptchaGT    string // 极验 GT
	CaptchaKey   string // 极验 Key
}

// GetLoginConfig 获取登录配置
func GetLoginConfig() *LoginConfig {
	conf, _ := config.GetConfigCached()
	return &LoginConfig{
		EnableCaptcha: conf.EnableCaptcha,
		CaptchaGT:    conf.CaptchaGT,
		CaptchaKey:   conf.CaptchaKey,
	}
}

// CaptchaInit 验证码初始化接口
func CaptchaInit(w http.ResponseWriter, r *http.Request) {
	loginCfg := GetLoginConfig()
	
	// 如果未启用验证码，返回成功但无数据
	if !loginCfg.EnableCaptcha || loginCfg.CaptchaGT == "" {
		returnOK(w, "Captcha disabled", nil)
		return
	}

	// 生成一个简单的 challenge 用于验证
	challenge := util.GenerateToken("captcha")
	
	// 返回极验初始化所需参数
	// 注意：实际生产环境需要调用极验服务端 SDK 生成
	returnData := map[string]interface{}{
		"gt":        loginCfg.CaptchaGT,
		"challenge": challenge,
		"new_captcha": true,
		"success":   1,
	}
	
	returnOK(w, "success", returnData)
}

// CaptchaVerify 验证码二次验证
func CaptchaVerify(result map[string]interface{}) bool {
	loginCfg := GetLoginConfig()
	
	// 如果未启用验证码，直接通过
	if !loginCfg.EnableCaptcha || loginCfg.CaptchaKey == "" {
		return true
	}

	// 实际生产环境需要调用极验服务端 SDK 验证
	// 这里仅作示例，实际需要：
	// 1. 从 result 中获取 geetest_challenge, geetest_validate, geetest_seccode
	// 2. 调用极验 API 进行二次验证
	// 
	// 示例验证逻辑：
	// if result["geetest_validate"] == "" || result["geetest_seccode"] == "" {
	//     return false
	// }
	// 调用极验服务端验证接口
	
	return true // 临时返回true，实际使用时替换为真实验证逻辑
}

// LoginFunc 登录接口（带验证码）
func LoginFunc(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept-Language")
	util.InitLogLang(accept)

	if ld.failedTimes >= 5 {
		loginUnlock()
		returnError(w, util.LogStr("登录失败次数过多，请稍后再试"))
		return
	}

	// 从请求中读取 JSON 数据
	var data struct {
		Username string                 `json:"Username"`
		Password string                `json:"Password"`
		Captcha  map[string]interface{} `json:"Captcha"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		returnError(w, err.Error())
		return
	}

	// 用户名密码不能为空
	if data.Username == "" || data.Password == "" {
		returnError(w, util.LogStr("必须输入用户名/密码"))
		return
	}

	// 验证码验证（如果启用）
	loginCfg := GetLoginConfig()
	if loginCfg.EnableCaptcha {
		if data.Captcha == nil {
			returnError(w, util.LogStr("请先完成滑动验证"))
			return
		}
		if !CaptchaVerify(data.Captcha) {
			returnError(w, util.LogStr("验证码验证失败，请重试"))
			return
		}
	}

	conf, _ := config.GetConfigCached()

	// 初始化用户名密码
	if conf.Username == "" && conf.Password == "" {
		if time.Since(startTime) > saveLimit {
			returnError(w, util.LogStr("需在 %s 之前完成用户名密码设置,请重启ddns-go", startTime.Add(saveLimit).Format("2006-01-02 15:04:05")))
			return
		}

		conf.NotAllowWanAccess = true
		u, err := url.Parse(r.Header.Get("referer"))
		if err == nil && !util.IsPrivateNetwork(u.Host) {
			conf.NotAllowWanAccess = false
		}

		conf.Username = data.Username
		hashedPwd, err := conf.CheckPassword(data.Password)
		if err != nil {
			returnError(w, err.Error())
			return
		}
		conf.Password = hashedPwd
		conf.SaveConfig()
	}

	// 登录
	if data.Username == conf.Username && util.PasswordOK(conf.Password, data.Password) {
		ld.ticker.Stop()
		ld.failedTimes = 0

		// 设置cookie过期时间为1天
		timeoutDays := 1
		if conf.NotAllowWanAccess {
			// 内网访问cookie过期时间为30天
			timeoutDays = 30
		}

		// 覆盖cookie
		cookieInSystem = &http.Cookie{
			Name:     cookieName,
			Value:    util.GenerateToken(data.Username), // 生成token
			Path:     "/",
			Expires:  time.Now().AddDate(0, 0, timeoutDays), // 设置过期时间
			HttpOnly: true,
		}
		// 写入cookie
		http.SetCookie(w, cookieInSystem)

		util.Log("%q 登录成功", util.GetRequestIPStr(r))

		returnOK(w, util.LogStr("登录成功"), cookieInSystem.Value)
		return
	}

	ld.failedTimes = ld.failedTimes + 1
	util.Log("%q 帐号密码不正确", util.GetRequestIPStr(r))
	returnError(w, util.LogStr("用户名或密码错误"))
}
