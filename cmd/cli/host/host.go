package host

import (
	"fmt"
	"time"
	
	"github.com/pion/webrtc/v3"
	"github.com/spf13/cobra"
	"github.com/submlit21/stardewl-ink/core"
)

var (
	modsPath string
)

var HostCmd = &cobra.Command{
	Use:   "host",
	Short: "Run as host (create a room)",
	Long: `Run in host mode to create a multiplayer room.

This command creates a new room on the signaling server and
generates a connection code for clients to join.

Examples:
  # Create a room with default settings
  stardewl host
  
  # Create a room with specific mods path
  stardewl host --mods /path/to/Mods
  
  # Create a room with 60-second timeout
  stardewl host --timeout 60`,
	Args: cobra.NoArgs,
	RunE: runHost,
}

func init() {
	HostCmd.Flags().StringVar(&modsPath, "mods", "", "Mods folder path (default: auto-detect)")
}

func runHost(cmd *cobra.Command, args []string) error {
	// Get global flags
	timeout, _ := cmd.Root().PersistentFlags().GetInt("timeout")
	signalingURL, _ := cmd.Root().PersistentFlags().GetString("signaling")
	
	fmt.Println("=== Host Mode ===")
	fmt.Printf("Signaling server: %s\n", signalingURL)
	fmt.Println("Creating room on signaling server...")
	
	// Create P2P connector configuration
	config := core.P2PConfig{
		SignalingURL: signalingURL,
		IsHost:       true,
		ModsPath:     modsPath,
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
			{URLs: []string{"stun:stun2.l.google.com:19302"}},
			{URLs: []string{"stun:stun3.l.google.com:19302"}},
			{URLs: []string{"stun:stun4.l.google.com:19302"}},
		},
	}
	
	// Auto-generate room ID
	roomID := core.GenerateRoomID()
	config.RoomID = roomID
	
	// Create P2P connector
	connector, err := core.NewP2PConnector(config)
	if err != nil {
		return fmt.Errorf("failed to create P2P connector: %v", err)
	}
	defer connector.Close()
	
	// Start connection
	if err := connector.Start(); err != nil {
		return fmt.Errorf("failed to start P2P connection: %v", err)
	}
	
	// Wait based on timeout setting
	if timeout > 0 {
		fmt.Printf("\nWaiting for %d seconds (timeout)...\n", timeout)
		time.Sleep(time.Duration(timeout) * time.Second)
		fmt.Println("Timeout reached, exiting...")
	} else {
		fmt.Print("\nPress Enter to exit...")
		var input string
		fmt.Scanln(&input)
	}
	
	return nil
}