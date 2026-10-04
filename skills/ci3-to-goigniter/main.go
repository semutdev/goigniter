package main

import (
	"fmt"
	"os"

	"github.com/semutdev/goigniter/skills/ci3-to-goigniter/scripts"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "inspect":
		if err := scripts.RunInspectCLI(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "sql2struct":
		if err := scripts.RunSQL2StructCLI(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "view2gotpl":
		if err := scripts.RunView2GoTplCLI(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("CI3 to GoIgniter Migration Toolkit")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run ./skills/ci3-to-goigniter <command> [flags]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  inspect     Scan a legacy CI3 project and produce an inventory report")
	fmt.Println("  sql2struct  Convert SQL schema / table to Go model structs (Task 3)")
	fmt.Println("  view2gotpl  Convert CI3 PHP views to Go html/template files (Task 4)")
	fmt.Println()
	fmt.Println("Use 'go run ./skills/ci3-to-goigniter <command> -h' for more information about a command.")
}
