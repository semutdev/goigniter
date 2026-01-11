package libs

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"goigniter/config"
	"goigniter/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Auth errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user is not activated")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrPasswordMismatch   = errors.New("current password is incorrect")
)

// JWTClaims untuk JWT token
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// SessionData untuk cookie session
type SessionData struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"`
}

// --- Core Auth Functions ---

// Login memverifikasi credentials dan return user
func Login(identity, password, ipAddress string) (*models.User, error) {
	var user models.User

	// Cari user by email atau username
	result := config.DB.Preload("Groups").
		Where("email = ? OR username = ?", identity, identity).
		First(&user)

	if result.Error != nil {
		// Log attempt
		logLoginAttempt(ipAddress, identity)
		return nil, ErrInvalidCredentials
	}

	// Cek password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logLoginAttempt(ipAddress, identity)
		return nil, ErrInvalidCredentials
	}

	// Cek active
	if !user.Active {
		return nil, ErrUserNotActive
	}

	// Update last login
	now := time.Now().Unix()
	user.LastLogin = &now
	user.IPAddress = ipAddress
	config.DB.Save(&user)

	return &user, nil
}

// RegisterUser membuat user baru
func RegisterUser(email, password, firstName, lastName, ipAddress string) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate activation code
	selector := generateToken(16)
	code := generateToken(32)
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)

	user := models.User{
		Email:              email,
		Password:           string(hashedPassword),
		FirstName:          &firstName,
		LastName:           &lastName,
		IPAddress:          ipAddress,
		CreatedOn:          time.Now().Unix(),
		Active:             false, // Butuh aktivasi
		ActivationSelector: &selector,
		ActivationCode:     stringPtr(string(hashedCode)),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	// Assign ke group members
	var membersGroup models.Group
	config.DB.Where("name = ?", "members").First(&membersGroup)
	config.DB.Model(&user).Association("Groups").Append(&membersGroup)

	return &user, nil
}

// Activate mengaktivasi user dengan code
func Activate(selector, code string) error {
	var user models.User
	result := config.DB.Where("activation_selector = ?", selector).First(&user)
	if result.Error != nil {
		return ErrInvalidToken
	}

	// Verify code
	if user.ActivationCode == nil {
		return ErrInvalidToken
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.ActivationCode), []byte(code)); err != nil {
		return ErrInvalidToken
	}

	// Activate user
	user.Active = true
	user.ActivationSelector = nil
	user.ActivationCode = nil
	config.DB.Save(&user)

	return nil
}

// ForgotPassword generates reset token dan return selector:code untuk email
func ForgotPassword(email string) (string, string, error) {
	var user models.User
	result := config.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return "", "", ErrUserNotFound
	}

	// Generate tokens
	selector := generateToken(16)
	code := generateToken(32)
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	now := time.Now().Unix()

	user.ForgottenPasswordSelector = &selector
	user.ForgottenPasswordCode = stringPtr(string(hashedCode))
	user.ForgottenPasswordTime = &now
	config.DB.Save(&user)

	return selector, code, nil
}

// ResetPassword mereset password dengan token
func ResetPassword(selector, code, newPassword string) error {
	var user models.User
	result := config.DB.Where("forgotten_password_selector = ?", selector).First(&user)
	if result.Error != nil {
		return ErrInvalidToken
	}

	// Cek expired (24 jam)
	if user.ForgottenPasswordTime == nil || time.Now().Unix()-*user.ForgottenPasswordTime > 86400 {
		return ErrInvalidToken
	}

	// Verify code
	if user.ForgottenPasswordCode == nil {
		return ErrInvalidToken
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.ForgottenPasswordCode), []byte(code)); err != nil {
		return ErrInvalidToken
	}

	// Update password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	user.ForgottenPasswordSelector = nil
	user.ForgottenPasswordCode = nil
	user.ForgottenPasswordTime = nil
	config.DB.Save(&user)

	return nil
}

// UpdatePassword mengupdate password user
func UpdatePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return ErrUserNotFound
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrPasswordMismatch
	}

	// Update password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	config.DB.Save(&user)

	return nil
}

// --- Session Functions (Web) ---

// SetSession menyimpan user session ke cookie
func SetSession(c echo.Context, user *models.User, remember bool) error {
	expiry := time.Now().Add(24 * time.Hour)
	if remember {
		expiry = time.Now().Add(30 * 24 * time.Hour) // 30 hari
	}

	sessionData := SessionData{
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: expiry.Unix(),
	}

	jsonData, _ := json.Marshal(sessionData)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	cookie := &http.Cookie{
		Name:     "goigniter_session",
		Value:    encoded,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   false, // Set true di production dengan HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return nil
}

// GetSession mengambil user dari session cookie
func GetSession(c echo.Context) *models.User {
	cookie, err := c.Cookie("goigniter_session")
	if err != nil {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil
	}

	var sessionData SessionData
	if err := json.Unmarshal(decoded, &sessionData); err != nil {
		return nil
	}

	// Cek expired
	if time.Now().Unix() > sessionData.ExpiresAt {
		return nil
	}

	var user models.User
	if err := config.DB.Preload("Groups").First(&user, sessionData.UserID).Error; err != nil {
		return nil
	}

	return &user
}

// ClearSession menghapus session cookie
func ClearSession(c echo.Context) {
	cookie := &http.Cookie{
		Name:     "goigniter_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	c.SetCookie(cookie)
}

// --- JWT Functions (API) ---

// GenerateJWT membuat JWT token
func GenerateJWT(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-this"
	}

	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT memvalidasi JWT token dan return user
func ValidateJWT(tokenString string) (*models.User, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-this"
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	var user models.User
	if err := config.DB.Preload("Groups").First(&user, claims.UserID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// --- Helper Functions ---

// GetUser mengambil user dari context (session atau JWT)
func GetUser(c echo.Context) *models.User {
	// Cek dari context dulu (sudah di-set middleware)
	if user, ok := c.Get("user").(*models.User); ok {
		return user
	}

	// Cek JWT dari header
	auth := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		if user, err := ValidateJWT(token); err == nil {
			return user
		}
	}

	// Cek session cookie
	return GetSession(c)
}

// IsLoggedIn cek apakah user sudah login
func IsLoggedIn(c echo.Context) bool {
	return GetUser(c) != nil
}

// InGroup cek apakah user ada di group tertentu
func InGroup(user *models.User, groupName string) bool {
	if user == nil {
		return false
	}
	for _, group := range user.Groups {
		if group.Name == groupName {
			return true
		}
	}
	return false
}

// RequireAuth cek login, redirect jika belum
func RequireAuth(c echo.Context) bool {
	if !IsLoggedIn(c) {
		c.Redirect(http.StatusFound, "/auth/login")
		return false
	}
	return true
}

// RequireGroup cek login + group membership
func RequireGroup(c echo.Context, groupName string) bool {
	user := GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/auth/login")
		return false
	}
	if !InGroup(user, groupName) {
		c.Redirect(http.StatusFound, "/auth/login?error=forbidden")
		return false
	}
	return true
}

// --- Internal Helpers ---

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func stringPtr(s string) *string {
	return &s
}

func logLoginAttempt(ipAddress, login string) {
	attempt := models.LoginAttempt{
		IPAddress: ipAddress,
		Login:     login,
		Time:      time.Now().Unix(),
	}
	config.DB.Create(&attempt)
}
