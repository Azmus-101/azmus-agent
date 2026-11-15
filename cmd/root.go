/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "azmus",
	Short: "Azmus agent: watch, understand and fix your infrastructure",
	Long: `Azmus is a lightweight agent that discovers what runs on your
servers and projects, and can act on it.

Run "azmus sniff <dir>" to scan a directory for infrastructure files.`,
}

func Execute() {
	// Cancel the context on Ctrl+C so long scans stop cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		stop() // os.Exit skips deferred calls, so release the signal handler first
		os.Exit(1)
	}
}
