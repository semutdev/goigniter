package controllers

import (
	"full-crud/application/libs"
	"net/http"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&AuthController{})
}

type AuthController struct {
	core.Controller
}

// LoginForm data
type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember string `json:"remember"`
}

// RegisterForm data
type RegisterForm struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
}

// Login menampilkan form login
func (a *AuthController) Login() {
	if libs.IsLoggedIn(a.Ctx) {
		a.Ctx.Redirect(http.StatusFound, "/")
		return
	}

	a.Ctx.View("auth/login", core.Map{
		"Title":  "Login",
		"Error":  a.Ctx.Query("error"),
		"Values": LoginForm{},
	})
}

// Dologin memproses login
func (a *AuthController) Dologin() {
	email := a.Ctx.FormValue("email")
	password := a.Ctx.FormValue("password")
	remember := a.Ctx.FormValue("remember")

	if email == "" || password == "" {
		a.Ctx.View("auth/login", core.Map{
			"Title": "Login",
			"Error": "Email dan password wajib diisi",
		})
		return
	}

	ipAddress := a.Ctx.IP()
	user, err := libs.Login(email, password, ipAddress)
	if err != nil {
		a.Ctx.View("auth/login", core.Map{
			"Title": "Login",
			"Error": err.Error(),
		})
		return
	}

	rememberBool := remember == "on" || remember == "1"
	libs.SetSession(a.Ctx, user, rememberBool)

	if libs.InGroup(user, "admin") {
		a.Ctx.Redirect(http.StatusFound, "/admin/dashboard")
		return
	}

	a.Ctx.Redirect(http.StatusFound, "/")
}

// Logout menghapus session
func (a *AuthController) Logout() {
	libs.ClearSession(a.Ctx)
	a.Ctx.Redirect(http.StatusFound, "/auth/login")
}

// Register menampilkan form register
func (a *AuthController) Register() {
	if libs.IsLoggedIn(a.Ctx) {
		a.Ctx.Redirect(http.StatusFound, "/")
		return
	}

	a.Ctx.View("auth/register", core.Map{
		"Title":  "Register",
		"Values": RegisterForm{},
		"Errors": map[string]string{},
	})
}

// Doregister memproses registrasi
func (a *AuthController) Doregister() {
	email := a.Ctx.FormValue("email")
	password := a.Ctx.FormValue("password")
	passwordConfirm := a.Ctx.FormValue("password_confirm")
	firstName := a.Ctx.FormValue("first_name")
	lastName := a.Ctx.FormValue("last_name")

	errors := make(map[string]string)
	if email == "" {
		errors["Email"] = "Email wajib diisi"
	}
	if password == "" {
		errors["Password"] = "Password wajib diisi"
	}
	if len(password) < 6 {
		errors["Password"] = "Password minimal 6 karakter"
	}
	if password != passwordConfirm {
		errors["PasswordConfirm"] = "Password tidak cocok"
	}

	if len(errors) > 0 {
		a.Ctx.View("auth/register", core.Map{
			"Title":  "Register",
			"Errors": errors,
			"Values": RegisterForm{
				Email:     email,
				FirstName: firstName,
				LastName:  lastName,
			},
		})
		return
	}

	ipAddress := a.Ctx.IP()
	_, err := libs.RegisterUser(email, password, firstName, lastName, ipAddress)
	if err != nil {
		a.Ctx.View("auth/register", core.Map{
			"Title":  "Register",
			"Error":  "Email sudah terdaftar",
			"Errors": map[string]string{},
		})
		return
	}

	a.Ctx.View("auth/register", core.Map{
		"Title":   "Register",
		"Success": "Registrasi berhasil! Silakan cek email untuk aktivasi.",
	})
}

// Activate mengaktivasi akun
func (a *AuthController) Activate() {
	selector := a.Ctx.Param("selector")
	code := a.Ctx.Param("code")

	if err := libs.Activate(selector, code); err != nil {
		a.Ctx.Redirect(http.StatusFound, "/auth/login?error=invalid_activation")
		return
	}

	a.Ctx.Redirect(http.StatusFound, "/auth/login?success=activated")
}

// Forgot menampilkan form forgot password
func (a *AuthController) Forgot() {
	a.Ctx.View("auth/forgot", core.Map{
		"Title": "Forgot Password",
	})
}

// Doforgot memproses forgot password
func (a *AuthController) Doforgot() {
	email := a.Ctx.FormValue("email")

	if email == "" {
		a.Ctx.View("auth/forgot", core.Map{
			"Title": "Forgot Password",
			"Error": "Email wajib diisi",
		})
		return
	}

	selector, code, err := libs.ForgotPassword(email)
	if err == nil {
		libs.SendForgotPasswordEmail(email, selector, code)
	}

	a.Ctx.View("auth/forgot", core.Map{
		"Title":   "Forgot Password",
		"Success": "Jika email terdaftar, kami akan mengirimkan link reset password.",
	})
}

// Reset menampilkan form reset password
func (a *AuthController) Reset() {
	selector := a.Ctx.Param("selector")
	code := a.Ctx.Param("code")

	a.Ctx.View("auth/reset", core.Map{
		"Title":    "Reset Password",
		"Selector": selector,
		"Code":     code,
	})
}

// Doreset memproses reset password
func (a *AuthController) Doreset() {
	selector := a.Ctx.Param("selector")
	code := a.Ctx.Param("code")
	password := a.Ctx.FormValue("password")
	passwordConfirm := a.Ctx.FormValue("password_confirm")

	if password != passwordConfirm || len(password) < 6 {
		a.Ctx.View("auth/reset", core.Map{
			"Title":    "Reset Password",
			"Error":    "Password minimal 6 karakter dan harus sama",
			"Selector": selector,
			"Code":     code,
		})
		return
	}

	if err := libs.ResetPassword(selector, code, password); err != nil {
		a.Ctx.Redirect(http.StatusFound, "/auth/login?error=invalid_reset")
		return
	}

	a.Ctx.Redirect(http.StatusFound, "/auth/login?success=password_reset")
}
