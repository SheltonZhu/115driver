package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/SheltonZhu/115driver/cli/internal/output"
	"github.com/SheltonZhu/115driver/cli/internal/resolver"
	"github.com/spf13/cobra"
)

var rmForce bool

var rmCmd = &cobra.Command{
	Use:   "rm <remote_path>...",
	Short: "Delete file or directory (moves to recycle bin)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileIDs := make([]string, len(args))
		directoryCount := 0
		for i, remotePath := range args {
			fileID, isDir, err := resolver.ResolvePath(client, remotePath)
			if err != nil {
				return &exitError{code: output.ExitNotFound, msg: err.Error()}
			}
			fileIDs[i] = fileID
			if isDir {
				directoryCount++
			}
		}

		if err := validateDeleteConfirmation(directoryCount > 0, jsonOutput, rmForce); err != nil {
			return &exitError{code: output.ExitArgs, msg: err.Error()}
		}

		if directoryCount > 0 && !jsonOutput && !rmForce {
			directoryWord := "directories"
			if directoryCount == 1 {
				directoryWord = "directory"
			}
			fmt.Printf("Delete %d item(s), including %d %s? [y/N] ", len(args), directoryCount, directoryWord)
			reader := bufio.NewReader(os.Stdin)
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "y" && resp != "yes" {
				fmt.Println("Canceled.")
				return nil
			}
		}

		if err := client.Delete(fileIDs...); err != nil {
			return &exitError{code: output.ExitError, msg: err.Error()}
		}

		printer.PrintSuccess(map[string]interface{}{
			"deleted":  args,
			"file_ids": fileIDs,
		})
		if !jsonOutput {
			for _, remotePath := range args {
				fmt.Printf("Deleted: %s\n", remotePath)
			}
		}
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Skip confirmation for directory deletes")
	rootCmd.AddCommand(rmCmd)
}

func validateDeleteConfirmation(isDir, jsonOutput, force bool) error {
	if !isDir {
		return nil
	}
	if force {
		return nil
	}
	if jsonOutput {
		return errors.New("directory delete requires --force when using --json")
	}
	return nil
}
