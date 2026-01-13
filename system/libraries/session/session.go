package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"goigniter/system/core"
)

// Config holds session configuration.
type Config struct {
	Secret     string        // HMAC signing key (required)
	CookieName string        // Cookie name (default: "goigniter_session")
	MaxAge     int           // Session max age in seconds (default: 86400)
	Path       string        // Cookie path (default: "/")
	Domain     string        // Cookie domain
	HttpOnly   bool          // HTTP only flag (default: true)
	Secure     bool          // Secure flag (default: false)
	SameSite   http.SameSite // SameSite policy (default: Lax)
}

// Session represents a user session.
type Session struct {
	ID        string            `json:"id"`
	Data      map[string]any    `json:"data"`
	Flash     map[string]string `json:"flash"`
	ExpiresAt int64             `json:"expires_at"`
}

var (
	config     Config
	configOnce sync.Once
)

// Init initializes the session with the given configuration.
func Init(cfg Config) {
	configOnce.Do(func() {
		config = cfg
		if config.CookieName == "" {
			config.CookieName = "goigniter_session"
		}
		if config.MaxAge == 0 {
			config.MaxAge = 86400 // 24 hours
		}
		if config.Path == "" {
			config.Path = "/"
		}
		if config.SameSite == 0 {
			config.SameSite = http.SameSiteLaxMode
		}
		config.HttpOnly = true // Always true for security
	})
}

// Get retrieves or creates a session for the current request.
func Get(c *core.Context) *Session {
	// Try to load from cookie
	cookie, err := c.Cookie(config.CookieName)
	if err == nil && cookie.Value != "" {
		session, err := decode(cookie.Value)
		if err == nil && session.ExpiresAt > time.Now().Unix() {
			return session
		}
	}

	// Create new session
	return &Session{
		ID:        generateID(),
		Data:      make(map[string]any),
		Flash:     make(map[string]string),
		ExpiresAt: time.Now().Unix() + int64(config.MaxAge),
	}
}

// Save saves the session to a cookie.
func (s *Session) Save(c *core.Context) error {
	// Update expiry
	s.ExpiresAt = time.Now().Unix() + int64(config.MaxAge)

	// Encode session
	value, err := encode(s)
	if err != nil {
		return err
	}

	// Set cookie
	cookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    value,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   config.MaxAge,
		HttpOnly: config.HttpOnly,
		Secure:   config.Secure,
		SameSite: config.SameSite,
	}
	c.SetCookie(cookie)

	return nil
}

// Set stores a value in the session.
func (s *Session) Set(key string, value any) {
	if s.Data == nil {
		s.Data = make(map[string]any)
	}
	s.Data[key] = value
}

// Get retrieves a value from the session.
func (s *Session) Get(key string) any {
	if s.Data == nil {
		return nil
	}
	return s.Data[key]
}

// GetString retrieves a string value from the session.
func (s *Session) GetString(key string) string {
	if v, ok := s.Data[key].(string); ok {
		return v
	}
	return ""
}

// GetInt retrieves an int value from the session.
func (s *Session) GetInt(key string) int {
	switch v := s.Data[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

// Delete removes a value from the session.
func (s *Session) Delete(key string) {
	delete(s.Data, key)
}

// Clear removes all data from the session.
func (s *Session) Clear() {
	s.Data = make(map[string]any)
	s.Flash = make(map[string]string)
}

// Destroy destroys the session by clearing the cookie.
func (s *Session) Destroy(c *core.Context) {
	cookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    "",
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: config.HttpOnly,
		Secure:   config.Secure,
		SameSite: config.SameSite,
	}
	c.SetCookie(cookie)
}

// --- Flash Messages ---

// SetFlash sets a flash message.
func SetFlash(c *core.Context, key, value string) {
	session := Get(c)
	if session.Flash == nil {
		session.Flash = make(map[string]string)
	}
	session.Flash[key] = value
	session.Save(c)
}

// GetFlash retrieves and removes a flash message.
func GetFlash(c *core.Context, key string) string {
	session := Get(c)
	if session.Flash == nil {
		return ""
	}

	value, exists := session.Flash[key]
	if !exists {
		return ""
	}

	// Remove flash after reading
	delete(session.Flash, key)
	session.Save(c)

	return value
}

// HasFlash checks if a flash message exists.
func HasFlash(c *core.Context, key string) bool {
	session := Get(c)
	if session.Flash == nil {
		return false
	}
	_, exists := session.Flash[key]
	return exists
}

// --- Internal Functions ---

// encode encodes and signs a session.
func encode(s *Session) (string, error) {
	// JSON encode
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	// Base64 encode
	encoded := base64.URLEncoding.EncodeToString(data)

	// Sign with HMAC
	signature := sign(encoded)

	// Return encoded.signature
	return encoded + "." + signature, nil
}

// decode decodes and verifies a session.
func decode(value string) (*Session, error) {
	// Split value into encoded and signature
	parts := splitOnce(value, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidSession
	}

	encoded, signature := parts[0], parts[1]

	// Verify signature
	if !verify(encoded, signature) {
		return nil, ErrInvalidSignature
	}

	// Base64 decode
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	// JSON decode
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// sign creates an HMAC signature.
func sign(data string) string {
	h := hmac.New(sha256.New, []byte(config.Secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// verify checks an HMAC signature.
func verify(data, signature string) bool {
	expected := sign(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// generateID generates a random session ID.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// splitOnce splits a string on the first occurrence of sep.
func splitOnce(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// Errors
var (
	ErrInvalidSession   = &SessionError{"invalid session"}
	ErrInvalidSignature = &SessionError{"invalid signature"}
)

type SessionError struct {
	Message string
}

func (e *SessionError) Error() string {
	return e.Message
}
