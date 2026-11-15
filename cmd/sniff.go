/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"os"

	"github.com/Azmus-101/azmus-agent/internal/sniffer"
	"github.com/spf13/cobra"
)

// sniffCmd represents the sniff command
var sniffCmd = &cobra.Command{
	Use:   "sniff [dir]",
	Short: "Scan a directory and print the infrastructure components found",
	Long: `Scans a directory for known infrastructure files (Dockerfiles, ...)
and prints the components it finds as JSON. Defaults to the current directory.

Example:
  azmus sniff test_data/projects/project1`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}

		snap, err := sniffer.New(sniffer.DefaultPlugins()).Run(cmd.Context(), os.DirFS(root))
		if err != nil {
			return err
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(snap)

	},
}

func init() {
	rootCmd.AddCommand(sniffCmd)
}
