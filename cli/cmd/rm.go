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
	Use:   "rm <remote_path>",
	Short: "Delete file or directory (moves to recycle bin)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]

		fileID, isDir, err := resolver.ResolvePath(client, remotePath)
		if err != nil {
			return &exitError{code: output.ExitNotFound, msg: err.Error()}
		}

		if err := validateDeleteConfirmation(isDir, jsonOutput, rmForce); err != nil {
			return &exitError{code: output.ExitArgs, msg: err.Error()}
		}

		if isDir && !jsonOutput && !rmForce {
			fmt.Printf("Delete directory %s and all its contents? [y/N] ", remotePath)
			reader := bufio.NewReader(os.Stdin)
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "y" && resp != "yes" {
				fmt.Println("Canceled.")
				return nil
			}
		}

		if err := client.Delete(fileID); err != nil {
			return &exitError{code: output.ExitError, msg: err.Error()}
		}

		printer.PrintSuccess(map[string]interface{}{
			"deleted":  []string{remotePath},
			"file_ids": []string{fileID},
		})
		if !jsonOutput {
			fmt.Printf("Deleted: %s\n", remotePath)
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
