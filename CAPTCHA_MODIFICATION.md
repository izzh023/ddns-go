# DDNS-GO 滑动验证码二次开发说明

## 修改文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/login_new.html` | 新增 | 新的登录页面（含验证码） |
| `web/login_new.go` | 新增 | 验证码初始化和验证接口 |
| `web/static/login.css` | 新增 | 美化后的登录样式 |
| `config/user.go` | 修改 | 添加验证码配置项 |

## 一、极验验证码申请步骤

### 1. 注册极验账号
访问 https://www.geetest.com 注册账号（个人开发者有免费额度）

### 2. 创建验证
登录后进入控制台 → 选择「行为验证」→ 创建验证 → 填写应用信息

### 3. 获取配置
创建后会获得：
- **Captcha ID** (gt) - 公开标识
- **Captcha Key** (key) - 私钥，用于服务端验证

### 4. 复制文件
```bash
# 备份原文件
mv web/login.html web/login.html.bak
mv web/login.go web/login.go.bak

# 使用新文件
cp web/login_new.html web/login.html
cp web/login_new.go web/login.go
```

## 二、配置验证码

### 方法1: 配置文件
编辑 `~/.ddns_go_config.yaml`，添加：
```yaml
username: your_username
password: your_hashed_password
enableCaptcha: true
captchaGT: your_gt_id        # 极验 Captcha ID
captchaKey: your_key         # 极验 Key
```

### 方法2: 首次安装向导
首次访问时会引导配置用户名密码和验证码（需在 `login_new.go` 中启用）

## 三、编译运行

```bash
cd ~/Desktop/ddns-go
make build
./ddns-go -s install
```

## 四、验证效果

1. 访问 http://localhost:9876
2. 应看到美化后的登录页面
3. 应显示滑动验证组件
4. 完成滑动验证后才能登录

## 五、注意事项

### 1. 验证码可能加载慢
如果极验 CDN 访问慢，可以：
- 替换为国内 CDN 镜像
- 或暂时禁用验证码（`enableCaptcha: false`）

### 2. 二次验证逻辑
当前 `CaptchaVerify` 函数中验证逻辑是示例代码，需要接入极验服务端 SDK：

```go
// 在 login_new.go 的 CaptchaVerify 函数中补充
// 参考极验官方文档：https://docs.geetest.com/kapps/overview/api/server
```

### 3. 关闭验证码
如果不想用验证码，修改配置：
```yaml
enableCaptcha: false
```

## 六、界面预览

### 美化前
- 简陋的用户名/密码表单
- 无任何保护

### 美化后
- 卡片式登录框，渐变背景
- 输入框图标提示
- 滑动验证组件
- 加载状态动画
- 暗色模式适配
