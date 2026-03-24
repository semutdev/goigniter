package libs

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"full-crud/application/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/libraries/database"
	"golang.org/x/crypto/bcrypt"
)

// Auth errors
var (
	ErrEmailNotFound      = errors.New("email not found")
	ErrInvalidCredentials = errors.New("invalid password")
	ErrUserNotActive      = errors.New("user is not activated")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrPasswordMismatch   = errors.New("current password is incorrect")
)

// JWTClaims for JWT token
type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// SessionData for cookie session
type SessionData struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"`
}

// --- Core Auth Functions ---

// Login verifies credentials and returns user
func Login(identity, password, ipAddress string) (*models.User, error) {
	var user models.User

	err := database.Table("users").
		WhereRaw("email = ?", identity).
		First(&user)
	if err != nil {
		logLoginAttempt(ipAddress, identity)
		return nil, ErrEmailNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logLoginAttempt(ipAddress, identity)
		return nil, ErrInvalidCredentials
	}

	if !user.Active {
		return nil, ErrUserNotActive
	}

	// Load user groups
	loadUserGroups(&user)

	// Update last login
	now := time.Now().Unix()
	database.Table("users").Where("id", user.ID).Update(map[string]any{
		"last_login": now,
		"ip_address": ipAddress,
	})
	user.LastLogin = &now
	user.IPAddress = ipAddress

	return &user, nil
}

// RegisterUser creates a new user
func RegisterUser(email, password, firstName, lastName, ipAddress string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	selector := generateToken(16)
	code := generateToken(32)
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)

	now := time.Now().Unix()

	// Insert user
	userID, err := database.Table("users").InsertGetId(map[string]any{
		"email":               email,
		"password":            string(hashedPassword),
		"first_name":          firstName,
		"last_name":           lastName,
		"ip_address":          ipAddress,
		"created_on":          now,
		"active":              0,
		"activation_selector": selector,
		"activation_code":     string(hashedCode),
	})
	if err != nil {
		return nil, err
	}

	// Get members group ID
	var membersGroupID int64
	database.Table("groups").Where("name", "members").Select("id").First(&membersGroupID)

	// Assign to members group
	database.Table("users_groups").Insert(map[string]any{
		"user_id":  userID,
		"group_id": membersGroupID,
	})

	return &models.User{
		ID:        userID,
		Email:     email,
		FirstName: &firstName,
		LastName:  &lastName,
	}, nil
}

// Activate activates a user with code
func Activate(selector, code string) error {
	var user models.User
	err := database.Table("users").
		Where("activation_selector", selector).
		First(&user)
	if err != nil {
		return ErrInvalidToken
	}

	if user.ActivationCode == nil {
		return ErrInvalidToken
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.ActivationCode), []byte(code)); err != nil {
		return ErrInvalidToken
	}

	database.Table("users").Where("id", user.ID).Update(map[string]any{
		"active":              1,
		"activation_selector": nil,
		"activation_code":     nil,
	})

	return nil
}

// ForgotPassword generates reset token
func ForgotPassword(email string) (string, string, error) {
	var user models.User
	err := database.Table("users").Where("email", email).First(&user)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	selector := generateToken(16)
	code := generateToken(32)
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	now := time.Now().Unix()

	database.Table("users").Where("id", user.ID).Update(map[string]any{
		"forgotten_password_selector": selector,
		"forgotten_password_code":     string(hashedCode),
		"forgotten_password_time":     now,
	})

	return selector, code, nil
}

// ResetPassword resets password with token
func ResetPassword(selector, code, newPassword string) error {
	var user models.User
	err := database.Table("users").
		Where("forgotten_password_selector", selector).
		First(&user)
	if err != nil {
		return ErrInvalidToken
	}

	if user.ForgottenPasswordTime == nil || time.Now().Unix()-*user.ForgottenPasswordTime > 86400 {
		return ErrInvalidToken
	}

	if user.ForgottenPasswordCode == nil {
		return ErrInvalidToken
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.ForgottenPasswordCode), []byte(code)); err != nil {
		return ErrInvalidToken
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	database.Table("users").Where("id", user.ID).Update(map[string]any{
		"password":                    string(hashedPassword),
		"forgotten_password_selector": nil,
		"forgotten_password_code":     nil,
		"forgotten_password_time":     nil,
	})

	return nil
}

// --- Session Functions ---

// SetSession stores user session in cookie
func SetSession(c *core.Context, user *models.User, remember bool) error {
	expiry := time.Now().Add(24 * time.Hour)
	if remember {
		expiry = time.Now().Add(30 * 24 * time.Hour)
	}

	sessionData := SessionData{
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: expiry.Unix(),
	}

	jsonData, _ := json.Marshal(sessionData)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	cookie := &http.Cookie{
		Name:     "goigniter_auth",
		Value:    encoded,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return nil
}

// GetSession retrieves user from session cookie
func GetSession(c *core.Context) *models.User {
	cookie, err := c.Cookie("goigniter_auth")
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

	if time.Now().Unix() > sessionData.ExpiresAt {
		return nil
	}

	var user models.User
	err = database.Table("users").Where("id", sessionData.UserID).First(&user)
	if err != nil {
		return nil
	}

	loadUserGroups(&user)
	return &user
}

// ClearSession removes session cookie
func ClearSession(c *core.Context) {
	cookie := &http.Cookie{
		Name:     "goigniter_auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	c.SetCookie(cookie)
}

// --- JWT Functions (API) ---

// GenerateJWT creates a JWT token
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

// ValidateJWT validates JWT token
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
	err = database.Table("users").Where("id", claims.UserID).First(&user)
	if err != nil {
		return nil, ErrUserNotFound
	}

	loadUserGroups(&user)
	return &user, nil
}

// --- Helper Functions ---

func loadUserGroups(user *models.User) {
	var groups []models.Group
	database.Query(`
		SELECT g.id, g.name, g.description
		FROM groups g
		INNER JOIN users_groups ug ON g.id = ug.group_id
		WHERE ug.user_id = ?
	`, user.ID).Get(&groups)
	user.Groups = groups
}

// GetUser retrieves user from context
func GetUser(c *core.Context) *models.User {
	if user, ok := c.Get("user").(*models.User); ok {
		return user
	}

	auth := c.Header("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		if user, err := ValidateJWT(token); err == nil {
			return user
		}
	}

	return GetSession(c)
}

// IsLoggedIn checks if user is logged in
func IsLoggedIn(c *core.Context) bool {
	return GetUser(c) != nil
}

// InGroup checks if user belongs to a group
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

// RequireAuth middleware - redirect if not logged in
func RequireAuth(c *core.Context) bool {
	if !IsLoggedIn(c) {
		c.Redirect(http.StatusFound, "/auth/login")
		return false
	}
	return true
}

// RequireGroup checks login + group membership
func RequireGroup(c *core.Context, groupName string) bool {
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

func logLoginAttempt(ipAddress, login string) {
	database.Table("login_attempts").Insert(map[string]any{
		"ip_address": ipAddress,
		"login":      login,
		"time":       time.Now().Unix(),
	})
}
