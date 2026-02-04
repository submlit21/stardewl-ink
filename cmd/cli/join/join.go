package join

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	
	"github.com/pion/webrtc/v3"
	"github.com/spf13/cobra"
	"github.com/submlit21/stardewl-ink/core"
)

var (
	modsPath string
)

var JoinCmd = &cobra.Command{
	Use:   "join [connection-code]",
	Short: "Run as client (join a room)",
	Long: `Run in client mode to join a multiplayer room.

This command connects to an existing room using the connection code
provided by the host.

Examples:
  # Join a room with connection code 123456
  stardewl join 123456
  
  # Join with specific mods path
  stardewl join 123456 --mods /path/to/Mods
  
  # Join with 30-second timeout
  stardewl join 123456 --timeout 30`,
	Args: cobra.ExactArgs(1),
	RunE: runJoin,
}

func init() {
	JoinCmd.Flags().StringVar(&modsPath, "mods", "", "Mods folder path (default: auto-detect)")
}

func runJoin(cmd *cobra.Command, args []string) error {
	connectionID := args[0]
	
	// Get global flags
	timeout, _ := cmd.Root().PersistentFlags().GetInt("timeout")
	signalingURL, _ := cmd.Root().PersistentFlags().GetString("signaling")
	
	fmt.Println("=== Client Mode ===")
	fmt.Printf("Connection code: %s\n", connectionID)
	fmt.Printf("Signaling server: %s\n", signalingURL)
	fmt.Println("Connecting to host...")
	fmt.Println("(Press Ctrl+C to exit)")
	
	// Verify room exists
	fmt.Println("Verifying room exists...")
	checkRoomURL := strings.Replace(signalingURL, "ws://", "http://", 1)
	checkRoomURL = strings.Replace(checkRoomURL, "/ws", "/join/"+connectionID, 1)
	
	resp, err := http.Get(checkRoomURL)
	if err != nil {
		fmt.Println("Please ensure signaling server is running: ./dist/stardewl-signaling")
		return fmt.Errorf("failed to connect to signaling server: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		fmt.Printf("❌ Room does not exist: %s\n", connectionID)
		fmt.Println("Please check connection code, or wait for host to create room")
		return fmt.Errorf("room not found")
	} else if resp.StatusCode != 200 {
		fmt.Printf("❌ Failed to verify room, status code: %d\n", resp.StatusCode)
		return fmt.Errorf("room verification failed")
	}
	
	var roomResponse struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Ready   bool   `json:"ready"`
		Message string `json:"message,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&roomResponse); err != nil {
		fmt.Printf("❌ Failed to parse room response: %v\n", err)
		return fmt.Errorf("failed to parse room response: %v", err)
	}
	
	if roomResponse.Ready {
		fmt.Println("✅ Room verified (host connected)")
	} else {
		fmt.Println("⚠️  Room exists but host not connected")
		fmt.Println("Please wait for host to connect, or check if host is running")
	}
	
	// Create P2P connector configuration
	config := core.P2PConfig{
		SignalingURL: signalingURL,
		RoomID:       connectionID,
		IsHost:       false,
		ModsPath:     modsPath,
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
			{URLs: []string{"stun:stun2.l.google.com:19302"}},
			{URLs: []string{"stun:stun3.l.google.com:19302"}},
			{URLs: []string{"stun:stun4.l.google.com:19302"}},
		},
	}
	
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
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
	
	return nil
}