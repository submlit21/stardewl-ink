package cli

import (
	"fmt"
	
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version number of Stardewl-Ink.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stardewl-Ink v0.1.0-alpha")
		fmt.Println("WebRTC P2P connection tool for Stardew Valley")
		fmt.Println("GitHub: https://github.com/submlit21/stardewl-ink")
	},
}