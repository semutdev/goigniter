package libs

import (
	"gopkg.in/gomail.v2"
	"os"
	"strconv"
)

// SendMail mengirim email menggunakan gomail
func SendMail(to, subject, body string) error {
	host := os.Getenv("MAIL_HOST")
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	user := os.Getenv("MAIL_USER")
	pass := os.Getenv("MAIL_PASS")
	from := os.Getenv("MAIL_FROM")

	if from == "" {
		from = user
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, user, pass)

	return d.DialAndSend(m)
}

// SendActivationEmail mengirim email aktivasi
func SendActivationEmail(email, selector, code string) error {
	link := BaseURL("/auth/activate/" + selector + "/" + code)
	body := `
		<h2>Aktivasi Akun</h2>
		<p>Klik link berikut untuk mengaktivasi akun Anda:</p>
		<p><a href="` + link + `">` + link + `</a></p>
	`
	return SendMail(email, "Aktivasi Akun", body)
}

// SendForgotPasswordEmail mengirim email reset password
func SendForgotPasswordEmail(email, selector, code string) error {
	link := BaseURL("/auth/reset/" + selector + "/" + code)
	body := `
		<h2>Reset Password</h2>
		<p>Klik link berikut untuk mereset password Anda:</p>
		<p><a href="` + link + `">` + link + `</a></p>
		<p>Link ini berlaku selama 24 jam.</p>
	`
	return SendMail(email, "Reset Password", body)
}
