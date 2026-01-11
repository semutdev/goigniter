package controllers

import (
	"goigniter/libs"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func init() {
	libs.Register("auth", &Auth{})
}

type Auth struct{}

// LoginForm menampilkan form login
type LoginForm struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required"`
	Remember string `form:"remember"`
}

// RegisterForm menampilkan form register
type RegisterForm struct {
	Email           string `form:"email" validate:"required,email"`
	Password        string `form:"password" validate:"required,min=6"`
	PasswordConfirm string `form:"password_confirm" validate:"required,eqfield=Password"`
	FirstName       string `form:"first_name" validate:"required"`
	LastName        string `form:"last_name" validate:"required"`
}

// ForgotForm form forgot password
type ForgotForm struct {
	Email string `form:"email" validate:"required,email"`
}

// ResetForm form reset password
type ResetForm struct {
	Password        string `form:"password" validate:"required,min=6"`
	PasswordConfirm string `form:"password_confirm" validate:"required,eqfield=Password"`
}

// Login menampilkan form login
func (a *Auth) Login(c echo.Context) error {
	// Jika sudah login, redirect ke home
	if libs.IsLoggedIn(c) {
		return c.Redirect(http.StatusFound, "/")
	}

	data := map[string]interface{}{
		"Title":  "Login",
		"Error":  c.QueryParam("error"),
		"Values": LoginForm{},
	}

	return c.Render(http.StatusOK, "auth/login", data)
}

// Dologin memproses login
func (a *Auth) Dologin(c echo.Context) error {
	form := new(LoginForm)

	if err := c.Bind(form); err != nil {
		return c.Redirect(http.StatusFound, "/auth/login?error=bad_request")
	}

	if err := c.Validate(form); err != nil {
		data := map[string]interface{}{
			"Title":  "Login",
			"Error":  "Validation failed",
			"Values": form,
		}
		return c.Render(http.StatusOK, "auth/login", data)
	}

	// Attempt login
	ipAddress := c.RealIP()
	user, err := libs.Login(form.Email, form.Password, ipAddress)
	if err != nil {
		data := map[string]interface{}{
			"Title":  "Login",
			"Error":  err.Error(),
			"Values": form,
		}
		return c.Render(http.StatusOK, "auth/login", data)
	}

	// Set session
	remember := form.Remember == "on" || form.Remember == "1"
	libs.SetSession(c, user, remember)

	// Redirect berdasarkan group
	if libs.InGroup(user, "admin") {
		return c.Redirect(http.StatusFound, "/admin/dashboard")
	}

	return c.Redirect(http.StatusFound, "/")
}

// Logout menghapus session
func (a *Auth) Logout(c echo.Context) error {
	libs.ClearSession(c)
	return c.Redirect(http.StatusFound, "/auth/login")
}

// Register menampilkan form register
func (a *Auth) Register(c echo.Context) error {
	if libs.IsLoggedIn(c) {
		return c.Redirect(http.StatusFound, "/")
	}

	data := map[string]interface{}{
		"Title":  "Register",
		"Values": RegisterForm{},
		"Errors": map[string]string{},
	}

	return c.Render(http.StatusOK, "auth/register", data)
}

// Doregister memproses registrasi
func (a *Auth) Doregister(c echo.Context) error {
	form := new(RegisterForm)

	if err := c.Bind(form); err != nil {
		return c.Redirect(http.StatusFound, "/auth/register?error=bad_request")
	}

	if err := c.Validate(form); err != nil {
		data := map[string]interface{}{
			"Title":  "Register",
			"Error":  "Validation failed",
			"Values": form,
			"Errors": getValidationErrors(err),
		}
		return c.Render(http.StatusOK, "auth/register", data)
	}

	// Register user
	ipAddress := c.RealIP()
	user, err := libs.RegisterUser(form.Email, form.Password, form.FirstName, form.LastName, ipAddress)
	if err != nil {
		data := map[string]interface{}{
			"Title":  "Register",
			"Error":  "Email sudah terdaftar",
			"Values": form,
			"Errors": map[string]string{},
		}
		return c.Render(http.StatusOK, "auth/register", data)
	}

	// Kirim email aktivasi (jika MAIL_USER diset)
	if user.ActivationSelector != nil && user.ActivationCode != nil {
		// Untuk development, tampilkan link aktivasi
		// Di production, kirim email
		// libs.SendActivationEmail(user.Email, *user.ActivationSelector, code)
	}

	data := map[string]interface{}{
		"Title":   "Register",
		"Success": "Registrasi berhasil! Silakan cek email untuk aktivasi.",
	}

	return c.Render(http.StatusOK, "auth/register", data)
}

// Activate mengaktivasi akun
func (a *Auth) Activate(c echo.Context) error {
	selector := c.Param("selector")
	code := c.Param("code")

	if err := libs.Activate(selector, code); err != nil {
		return c.Redirect(http.StatusFound, "/auth/login?error=invalid_activation")
	}

	return c.Redirect(http.StatusFound, "/auth/login?success=activated")
}

// Forgot menampilkan form forgot password
func (a *Auth) Forgot(c echo.Context) error {
	data := map[string]interface{}{
		"Title":  "Forgot Password",
		"Values": ForgotForm{},
	}

	return c.Render(http.StatusOK, "auth/forgot", data)
}

// Doforgot memproses forgot password
func (a *Auth) Doforgot(c echo.Context) error {
	form := new(ForgotForm)

	if err := c.Bind(form); err != nil {
		return c.Redirect(http.StatusFound, "/auth/forgot?error=bad_request")
	}

	if err := c.Validate(form); err != nil {
		data := map[string]interface{}{
			"Title":  "Forgot Password",
			"Error":  "Email tidak valid",
			"Values": form,
		}
		return c.Render(http.StatusOK, "auth/forgot", data)
	}

	selector, code, err := libs.ForgotPassword(form.Email)
	if err != nil {
		// Jangan kasih tahu user tidak ada (security)
		data := map[string]interface{}{
			"Title":   "Forgot Password",
			"Success": "Jika email terdaftar, kami akan mengirimkan link reset password.",
		}
		return c.Render(http.StatusOK, "auth/forgot", data)
	}

	// Kirim email
	libs.SendForgotPasswordEmail(form.Email, selector, code)

	data := map[string]interface{}{
		"Title":   "Forgot Password",
		"Success": "Link reset password telah dikirim ke email Anda.",
	}

	return c.Render(http.StatusOK, "auth/forgot", data)
}

// Reset menampilkan form reset password
func (a *Auth) Reset(c echo.Context) error {
	selector := c.Param("selector")
	code := c.Param("code")

	data := map[string]interface{}{
		"Title":    "Reset Password",
		"Selector": selector,
		"Code":     code,
		"Values":   ResetForm{},
	}

	return c.Render(http.StatusOK, "auth/reset", data)
}

// Doreset memproses reset password
func (a *Auth) Doreset(c echo.Context) error {
	selector := c.Param("selector")
	code := c.Param("code")

	form := new(ResetForm)

	if err := c.Bind(form); err != nil {
		return c.Redirect(http.StatusFound, "/auth/login?error=bad_request")
	}

	if err := c.Validate(form); err != nil {
		data := map[string]interface{}{
			"Title":    "Reset Password",
			"Error":    "Password minimal 6 karakter dan harus sama",
			"Selector": selector,
			"Code":     code,
			"Values":   form,
		}
		return c.Render(http.StatusOK, "auth/reset", data)
	}

	if err := libs.ResetPassword(selector, code, form.Password); err != nil {
		return c.Redirect(http.StatusFound, "/auth/login?error=invalid_reset")
	}

	return c.Redirect(http.StatusFound, "/auth/login?success=password_reset")
}

func getValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			// Custom pesan error ala CI3
			switch fieldError.Tag() {
			case "required":
				errors[fieldError.Field()] = "Field ini wajib diisi bro"
			case "email":
				errors[fieldError.Field()] = "Format email salah"
			case "min":
				errors[fieldError.Field()] = "Minimal " + fieldError.Param() + " karakter"
			default:
				errors[fieldError.Field()] = "Error validasi"
			}
		}
	}
	return errors
}
