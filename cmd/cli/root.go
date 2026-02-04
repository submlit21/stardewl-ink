package cli

import (
	"github.com/spf13/cobra"
	"github.com/submlit21/stardewl-ink/cmd/cli/host"
	"github.com/submlit21/stardewl-ink/cmd/cli/join"
	"github.com/submlit21/stardewl-ink/cmd/cli/mods"
	"github.com/submlit21/stardewl-ink/cmd/cli/signaling"
)

var (
	verbose    bool
	timeout    int
	signalingURL string
)

var rootCmd = &cobra.Command{
	Use:   "stardewl",
	Short: "Stardew Valley multiplayer tool using WebRTC P2P connections",
	Long: `Stardewl-Ink - Stardew Valley multiplayer tool

A WebRTC-based P2P connection tool for Stardew Valley multiplayer.
No port forwarding required, uses connection codes for pairing.

Examples:
  # Run as host (create a room)
  stardewl host
  
  # Run as client (join a room)
  stardewl join 123456
  
  # Run signaling server
  stardewl signaling
  
  # List mods
  stardewl mods list
  
  # Check mods in specific path
  stardewl mods list --path /path/to/Mods`,
	Version: "0.1.0-alpha",
	SilenceUsage: true,
	SilenceErrors: true,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 0, "Timeout in seconds (0 = wait indefinitely)")
	rootCmd.PersistentFlags().StringVar(&signalingURL, "signaling", "ws://localhost:8080/ws", "Signaling server URL")
	
	// Add subcommands
	rootCmd.AddCommand(host.HostCmd)
	rootCmd.AddCommand(join.JoinCmd)
	rootCmd.AddCommand(signaling.SignalingCmd)
	rootCmd.AddCommand(mods.ModsCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}