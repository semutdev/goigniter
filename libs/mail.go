package libs

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// MailConfig konfigurasi SMTP
type MailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	FromName string
}

// GetMailConfig mengambil konfigurasi dari environment
func GetMailConfig() MailConfig {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if port == 0 {
		port = 587
	}

	return MailConfig{
		Host:     getEnvDefault("MAIL_HOST", "smtp.gmail.com"),
		Port:     port,
		User:     os.Getenv("MAIL_USER"),
		Password: os.Getenv("MAIL_PASSWORD"),
		From:     getEnvDefault("MAIL_FROM", "noreply@goigniter.com"),
		FromName: getEnvDefault("MAIL_FROM_NAME", "GoIgniter"),
	}
}

// SendMail mengirim email
func SendMail(to, subject, body string) error {
	config := GetMailConfig()

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", config.FromName, config.From))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(config.Host, config.Port, config.User, config.Password)

	return d.DialAndSend(m)
}

// SendActivationEmail mengirim email aktivasi
func SendActivationEmail(to, selector, code string) error {
	appURL := getEnvDefault("APP_URL", "http://localhost:6789")
	activationLink := fmt.Sprintf("%s/auth/activate/%s/%s", appURL, selector, code)

	body := fmt.Sprintf(`
		<h2>Aktivasi Akun GoIgniter</h2>
		<p>Terima kasih telah mendaftar. Klik link berikut untuk mengaktivasi akun Anda:</p>
		<p><a href="%s">%s</a></p>
		<p>Link ini akan kadaluarsa dalam 24 jam.</p>
		<p>Jika Anda tidak mendaftar, abaikan email ini.</p>
	`, activationLink, activationLink)

	return SendMail(to, "Aktivasi Akun GoIgniter", body)
}

// SendForgotPasswordEmail mengirim email reset password
func SendForgotPasswordEmail(to, selector, code string) error {
	appURL := getEnvDefault("APP_URL", "http://localhost:6789")
	resetLink := fmt.Sprintf("%s/auth/reset/%s/%s", appURL, selector, code)

	body := fmt.Sprintf(`
		<h2>Reset Password GoIgniter</h2>
		<p>Anda menerima email ini karena ada permintaan reset password untuk akun Anda.</p>
		<p>Klik link berikut untuk reset password:</p>
		<p><a href="%s">%s</a></p>
		<p>Link ini akan kadaluarsa dalam 24 jam.</p>
		<p>Jika Anda tidak meminta reset password, abaikan email ini.</p>
	`, resetLink, resetLink)

	return SendMail(to, "Reset Password GoIgniter", body)
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
