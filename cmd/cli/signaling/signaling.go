package signaling

import (
	"fmt"
	"os"
	"os/exec"
	
	"github.com/spf13/cobra"
)

var (
	port string
)

var SignalingCmd = &cobra.Command{
	Use:   "signaling",
	Short: "Run the signaling server",
	Long: `Run the WebSocket signaling server required for WebRTC handshake.

The signaling server facilitates the initial connection between
host and client before direct P2P connection is established.

Examples:
  # Run on default port (8080)
  stardewl signaling
  
  # Run on specific port
  stardewl signaling --port 9090`,
	Args: cobra.NoArgs,
	RunE: runSignaling,
}

func init() {
	SignalingCmd.Flags().StringVar(&port, "port", ":8080", "Port to listen on")
}

func runSignaling(cmd *cobra.Command, args []string) error {
	fmt.Printf("Starting signaling server on port %s\n", port)
	fmt.Println("Note: The signaling server is a separate executable")
	fmt.Println("Building and running stardewl-signaling...")
	
	// Build the signaling server if not exists
	if _, err := os.Stat("./dist/stardewl-signaling"); os.IsNotExist(err) {
		fmt.Println("Building signaling server...")
		buildCmd := exec.Command("go", "build", "-o", "./dist/stardewl-signaling", "./signaling")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("failed to build signaling server: %v", err)
		}
	}
	
	// Run the signaling server
	fmt.Println("Starting signaling server...")
	runCmd := exec.Command("./dist/stardewl-signaling")
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	
	// Set environment variable for port if specified
	if port != ":8080" {
		runCmd.Env = append(os.Environ(), "PORT="+port)
	}
	
	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("signaling server exited with error: %v", err)
	}
	
	return nil
}