package core

// Version information - single source of truth
const (
	Version = "0.2.0"
	Website = "https://goigniter.semut.dev"
	Github  = "https://github.com/semutdev/goigniter"
)

// GetVersion returns the current framework version.
func GetVersion() string {
	return Version
}