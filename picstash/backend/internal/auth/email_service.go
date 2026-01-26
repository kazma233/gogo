package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log/slog"
	"math/big"
	"net/smtp"
	"path/filepath"

	"picstash/internal/config"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
	fromName string
	isDev    bool
}

type VerificationEmailData struct {
	Code      string
	ExpiresIn int
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		host:     cfg.SMTP.Host,
		port:     cfg.SMTP.Port,
		username: cfg.SMTP.Username,
		password: cfg.SMTP.Password,
		from:     cfg.SMTP.From,
		fromName: cfg.SMTP.FromName,
		isDev:    cfg.Server.Mode == "debug",
	}
}

func (e *EmailService) SendVerificationCode(to, code string, expiresIn int) error {
	if e.isDev {
		slog.Info("[开发模式] 验证码已记录", "to", to, "code", code, "expires_in", expiresIn)
		return nil
	}

	tmplPath, err := filepath.Abs("../../templates/emails/verification.html")
	if err != nil {
		return fmt.Errorf("获取模板路径失败: %w", err)
	}

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("解析邮件模板失败: %w", err)
	}

	data := VerificationEmailData{
		Code:      code,
		ExpiresIn: expiresIn,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("执行邮件模板失败: %w", err)
	}

	htmlBody := buf.String()

	msg := e.buildHTMLMail(to, "验证码", htmlBody, code)

	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	addr := fmt.Sprintf("%s:%s", e.host, e.port)

	if err := smtp.SendMail(addr, auth, e.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

func (e *EmailService) buildHTMLMail(to, subject, htmlBody, code string) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "From: %s <%s>\r\n", e.fromName, e.from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")

	boundary := "boundary_" + randString(16)
	fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	fmt.Fprintf(&buf, "\r\n")

	plainText := "您的验证码是：" + code
	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=utf-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n\r\n")
	buf.WriteString(base64.StdEncoding.EncodeToString([]byte(plainText)) + "\r\n")

	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	fmt.Fprintf(&buf, "Content-Type: text/html; charset=utf-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n\r\n")
	buf.WriteString(base64.StdEncoding.EncodeToString([]byte(htmlBody)) + "\r\n")

	fmt.Fprintf(&buf, "--%s--\r\n", boundary)

	return buf.String()
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[nBig.Int64()]
	}
	return string(b)
}

func GenerateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}
