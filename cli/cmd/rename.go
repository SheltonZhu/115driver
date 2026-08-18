package cmd

import (
	"fmt"
	"path"

	"github.com/SheltonZhu/115driver/cli/internal/output"
	"github.com/SheltonZhu/115driver/cli/internal/resolver"
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <remote_path> <new_name> [<remote_path> <new_name>...]",
	Short: "Rename files or directories",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 || len(args)%2 != 0 {
			return fmt.Errorf("requires one or more <remote_path> <new_name> pairs")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		renames, err := resolveRenames(args)
		if err != nil {
			return err
		}
		items := make([]driver.RenameItem, len(renames))
		for i, rename := range renames {
			items[i] = driver.RenameItem{FileID: rename.FileID, NewName: rename.NewName}
		}

		if err := client.BatchRename(items...); err != nil {
			return &exitError{code: output.ExitError, msg: err.Error()}
		}

		if len(renames) == 1 {
			printer.PrintSuccess(renames[0])
		} else {
			printer.PrintSuccess(map[string]interface{}{"renames": renames})
		}
		if !jsonOutput {
			for _, rename := range renames {
				fmt.Printf("Renamed %s -> %s\n", rename.RemotePath, rename.NewName)
			}
		}
		return nil
	},
}

type renameResult struct {
	RemotePath string `json:"remote_path"`
	OldName    string `json:"old_name"`
	NewName    string `json:"new_name"`
	FileID     string `json:"file_id"`
}

func resolveRenames(args []string) ([]renameResult, error) {
	renames := make([]renameResult, 0, len(args)/2)
	seen := make(map[string]struct{}, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		remotePath, newName := args[i], args[i+1]
		fileID, _, err := resolver.ResolvePath(client, remotePath)
		if err != nil {
			return nil, &exitError{code: output.ExitNotFound, msg: err.Error()}
		}
		if _, ok := seen[fileID]; ok {
			return nil, &exitError{code: output.ExitArgs, msg: fmt.Sprintf("duplicate file: %s", remotePath)}
		}
		seen[fileID] = struct{}{}
		renames = append(renames, renameResult{
			RemotePath: remotePath,
			OldName:    path.Base(remotePath),
			NewName:    newName,
			FileID:     fileID,
		})
	}
	return renames, nil
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
