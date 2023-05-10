package v1

import (
	"gopkg.in/gomail.v2"
)

var (
	MAIL_HOST = "smtp.qq.com" // 邮件服务器地址
	MAIL_PORT = 465           // 端口
	MAIL_USER = ""            // 发送邮件用户账号
	MAIL_PWD  = ""            // 授权密码
)

const EmailContentTemplate_RESET = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Forgot Password Verification Code</title>
</head>
<body style="font-family: Arial, sans-serif;">

    <h1>Forgot Password Verification Code</h1>

    <p>Dear ,</p>

    <p>You are requesting to reset your password. Please use the following verification code to proceed:</p>

    <p style="font-size: 24px; font-weight: bold;">Verification Code: %s (Valid for 5 minutes)</p>

    <p>If you didn't request this action, please ignore this email.</p>

    <p>https://mopdstor.com</p>

</body>
</html>
`

func SendGoMail(mailAddress []string, subject string, body string) error {
	m := gomail.NewMessage()
	// 这种方式可以添加别名，即 nickname， 也可以直接用<code>m.SetHeader("From", MAIL_USER)</code>
	nickname := `mopdstor | support`
	m.SetHeader("From", nickname+"<"+MAIL_USER+">")
	// 发送给多个用户
	m.SetHeader("To", mailAddress...)
	// 设置邮件主题
	m.SetHeader("Subject", subject)
	// 设置邮件正文
	m.SetBody(`text/html`, body)
	d := gomail.NewDialer(MAIL_HOST, MAIL_PORT, MAIL_USER, MAIL_PWD)
	// 发送邮件
	err := d.DialAndSend(m)
	return err
}
