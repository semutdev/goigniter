package core

import (
	"fmt"
	"runtime"
)

// Banner prints the GoIgniter startup banner with ASCII art logo.
// Call this at application startup for a nice visual.
func Banner(appName ...string) {
	name := "GoIgniter"
	if len(appName) > 0 && appName[0] != "" {
		name = appName[0]
	}

	// ASCII Art Logo
	logo := `
                   ++
                   + +
                  +   +
                ++    +
              ++     +++++
             ++     +++* ++
            ++            ++
           ++         +    +
           ++        +++   +
            ++     ++  +  ++
             ++       ++++
                +++ *+*+
                   +
`

	// Color codes
	cyan := "\033[36m"
	green := "\033[32m"
	yellow := "\033[33m"
	blue := "\033[34m"
	bold := "\033[1m"
	reset := "\033[0m"

	// Print banner
	fmt.Printf("%s%s", cyan, logo)
	fmt.Printf("%s%s%s v%s%s\n", bold, green, name, Version, reset)
	fmt.Printf("%s⚡ High Performance Web Framework for Go%s\n", yellow, reset)
	fmt.Printf("%s%s%s\n", blue, Website, reset)
	fmt.Println()

	// Print system info
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	fmt.Printf("  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
}

// BannerMini prints a minimal one-line banner.
func BannerMini(appName ...string) {
	name := "GoIgniter"
	if len(appName) > 0 && appName[0] != "" {
		name = appName[0]
	}

	cyan := "\033[36m"
	green := "\033[32m"
	bold := "\033[1m"
	reset := "\033[0m"

	fmt.Printf("%s%s%s%s v%s%s - Ready to serve! 🚀%s\n", bold, cyan, name, green, Version, reset, reset)
}
