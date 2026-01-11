# GoIgniter Auth System Design

Tanggal: 2026-01-11

## Tujuan

Implementasi authentication system ala Ion Auth 3 untuk GoIgniter.

## Keputusan Design

| Aspek | Keputusan |
|-------|-----------|
| Session/Token | Hybrid - JWT untuk API, Session untuk Web |
| Email | SMTP langsung (gomail) |
| Middleware | Role-based via group, cek di controller |
| Password | bcrypt hashing |

## Models

### User
```go
type User struct {
    ID                        uint
    IPAddress                 string
    Username                  string
    Password                  string
    Email                     string (unique)
    ActivationSelector        *string
    ActivationCode            *string
    ForgottenPasswordSelector *string
    ForgottenPasswordCode     *string
    ForgottenPasswordTime     *int64
    RememberSelector          *string
    RememberCode              *string
    CreatedOn                 int64
    LastLogin                 *int64
    Active                    bool
    FirstName                 *string
    LastName                  *string
    Company                   *string
    Phone                     *string
    Groups                    []Group (many2many)
}
```

### Group
```go
type Group struct {
    ID          uint
    Name        string (unique)
    Description string
}
```

### LoginAttempt
```go
type LoginAttempt struct {
    ID        uint
    IPAddress string
    Login     string
    Time      int64
}
```

## Library Auth

```go
// libs/auth.go
type AuthLib struct{}

// Core
func Login(identity, password string, remember bool) (*User, error)
func Logout(c echo.Context) error
func Register(email, password, firstName, lastName string) (*User, error)
func Activate(selector, code string) error
func ForgotPassword(email string) (string, error)
func ResetPassword(selector, code, newPassword string) error
func UpdatePassword(userID uint, oldPassword, newPassword string) error

// Helpers
func GetUser(c echo.Context) *User
func IsLoggedIn(c echo.Context) bool
func InGroup(user *User, groupName string) bool
func RequireAuth(c echo.Context) bool
func RequireGroup(c echo.Context, groupName string) bool

// JWT (API)
func GenerateJWT(user *User) (string, error)
func ValidateJWT(token string) (*User, error)

// Session (Web)
func SetSession(c echo.Context, user *User) error
func GetSession(c echo.Context) *User
func ClearSession(c echo.Context) error
```

## Controller Auth

| URL | Method | Action |
|-----|--------|--------|
| `/auth/login` | GET | Form login |
| `/auth/dologin` | POST | Proses login |
| `/auth/logout` | GET | Logout |
| `/auth/register` | GET | Form register |
| `/auth/doregister` | POST | Proses register |
| `/auth/forgot` | GET | Form forgot password |
| `/auth/doforgot` | POST | Kirim email reset |
| `/auth/reset/:code` | GET | Form reset password |
| `/auth/doreset/:code` | POST | Proses reset |

## Seeder

Default data:
- Groups: admin, members
- User: admin@admin.com / password (member of admin & members)

## Environment Variables

```env
# JWT
JWT_SECRET="your-secret-key"
JWT_EXPIRY="24h"

# Email SMTP
MAIL_HOST="smtp.gmail.com"
MAIL_PORT="587"
MAIL_USER="your@gmail.com"
MAIL_PASSWORD="app-password"
MAIL_FROM="noreply@goigniter.com"
MAIL_FROM_NAME="GoIgniter"

# Seeder
DB_SEED="false"
```

## Penggunaan di Controller

```go
func (d *Dashboard) Index(c echo.Context) error {
    // Require login + admin group
    if !libs.RequireGroup(c, "admin") {
        return c.Redirect(302, "/auth/login")
    }

    user := libs.GetUser(c)
    // ...
}
```
