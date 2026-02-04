package mods

import (
	"fmt"
	
	"github.com/spf13/cobra"
	"github.com/submlit21/stardewl-ink/core"
)

var (
	modsPath string
)

var ModsCmd = &cobra.Command{
	Use:   "mods",
	Short: "Manage Mods",
	Long: `Manage and list Stardew Valley mods.

This command helps you scan and list mods in your Stardew Valley
Mods folder for compatibility checking.

Examples:
  # List mods in auto-detected path
  stardewl mods list
  
  # List mods in specific path
  stardewl mods list --path /path/to/Mods`,
	Args: cobra.NoArgs,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List mods in Mods folder",
	Long:  `Scan and list all mods found in the Stardew Valley Mods folder.`,
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&modsPath, "path", "", "Mods folder path (default: auto-detect)")
	ModsCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	fmt.Println("=== Listing Mods ===")
	
	mods, err := core.ScanMods(modsPath)
	if err != nil {
		return fmt.Errorf("failed to scan mods: %v", err)
	}
	
	if len(mods) == 0 {
		fmt.Println("No mods found.")
		if modsPath == "" {
			fmt.Println("Try specifying the path with --path flag")
		}
		return nil
	}
	
	fmt.Printf("Found %d mods:\n", len(mods))
	for _, mod := range mods {
		fmt.Printf("  • %s", mod.Name)
		if mod.Version != "" {
			fmt.Printf(" v%s", mod.Version)
		}
		fmt.Println()
		if mod.Path != "" {
			fmt.Printf("    Path: %s\n", mod.Path)
		}
		fmt.Printf("    Size: %d bytes, Checksum: %s\n", mod.Size, mod.Checksum[:8])
		fmt.Println()
	}
	
	return nil
}