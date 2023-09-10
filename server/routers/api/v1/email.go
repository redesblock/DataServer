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

const EmailContentTemplate_ORDER = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Order Payment Confirmation</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 20px;
        }
        .container {
            background-color: #ffffff;
            border-radius: 5px;
            box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
            padding: 20px;
        }
        h1 {
            color: #333;
        }
        p {
            font-size: 16px;
            line-height: 1.6;
            color: #555;
        }
        .footer {
            margin-top: 20px;
            font-size: 12px;
            color: #888;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Thank You for Your Successful Payment!</h1>
        <p>Dear %s,</p>
        <p>Thank you for your successful payment. Here are the details of your order:</p>
        <ul>
            <li><strong>Order Number:</strong> %s</li>
            <li><strong>Payment Amount:</strong> %s</li>
            <li><strong>Payment method:</strong> %s</li>
			<li><strong>Payment Date and Time:</strong> %s</li>
        </ul>
        <p>Your order has been successfully paid, and we will process it and arrange for delivery as soon as possible. If you have any questions or need further assistance, please feel free to contact our customer support team.</p>
        <p>Thank you for shopping with us, and we look forward to providing you with more quality products and services in the future!</p>
        <p>Wishing you a great day!</p>
        <div class="footer">
            <p>Warmest regards,</p>

            <p>Website: https://mopdstor.com</p>
            <p>Email: respond@monopro.io</p>
        </div>
    </div>
</body>
</html>

`

func SendGoMail(mailAddress []string, subject string, body string) error {
	m := gomail.NewMessage()
	// 这种方式可以添加别名，即 nickname， 也可以直接用<code>m.SetHeader("From", MAIL_USER)</code>
	nickname := `mopdstor`
	m.SetHeader("From", nickname+"<no-reply@mopdstor.com>")
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
