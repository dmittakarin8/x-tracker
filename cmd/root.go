package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "x-tracker",
	Short: "A CLI tool to track X (Twitter) following changes",
	Long: `x-tracker is a command-line tool that monitors X (Twitter) accounts
and tracks their following changes in real-time. It supports Discord webhook
notifications and provides an interactive terminal user interface.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 