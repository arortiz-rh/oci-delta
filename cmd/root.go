package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Use:   "oci-delta",
	Short: "Create and apply OCI image deltas",
	Long: `oci-delta is a tool for creating and applying deltas between OCI images.
It supports creating efficient delta images, applying deltas to reconstruct full images,
and importing delta images directly into container storage.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here if needed
}

// Root returns the root cobra command for use by documentation generators.
func Root() *cobra.Command {
	return rootCmd
}

// Logger interface for command output
type Logger interface {
	Debug(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Default(format string, args ...interface{})
}

// cmdLogger implements the Logger interface
type cmdLogger struct {
	debug bool
	quiet bool
}

func (l *cmdLogger) Default(format string, args ...interface{}) {
	// Debug is a much more detailed version of the default logs, no need for both to exist at the same time.
	if !l.quiet && !l.debug {
		fmt.Printf(format+"\n", args...)
	}
}

func (l *cmdLogger) Debug(format string, args ...interface{}) {
	if l.debug {
		fmt.Printf(format+"\n", args...)
	}
}

func (l *cmdLogger) Warning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
}
