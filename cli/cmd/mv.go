package cmd

import (
	"fmt"

	"github.com/SheltonZhu/115driver/cli/internal/output"
	"github.com/SheltonZhu/115driver/cli/internal/resolver"
	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv <source_path>... <destination_dir>",
	Short: "Move files into a destination directory",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires one or more <source_path> arguments and a <destination_dir>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return moveOrCopy(args[:len(args)-1], args[len(args)-1], client.Move)
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}

type transferFunc func(dirID string, fileIDs ...string) error

func moveOrCopy(srcPaths []string, dstDir string, fn transferFunc) error {
	fileIDs := make([]string, len(srcPaths))
	for i, srcPath := range srcPaths {
		fileID, _, err := resolver.ResolvePath(client, srcPath)
		if err != nil {
			return &exitError{code: output.ExitNotFound, msg: err.Error()}
		}
		fileIDs[i] = fileID
	}
	dirID, err := resolver.ResolveDir(client, dstDir)
	if err != nil {
		return &exitError{code: output.ExitNotFound, msg: fmt.Sprintf("Destination directory not found: %s", dstDir)}
	}

	if err := fn(dirID, fileIDs...); err != nil {
		return &exitError{code: output.ExitError, msg: err.Error()}
	}

	result := map[string]interface{}{
		"destination_dir": dstDir,
		"file_ids":        fileIDs,
	}
	if len(srcPaths) == 1 {
		result["source"] = srcPaths[0]
	} else {
		result["sources"] = srcPaths
	}
	printer.PrintSuccess(result)
	if !jsonOutput {
		for _, srcPath := range srcPaths {
			fmt.Printf("Transferred %s -> %s\n", srcPath, dstDir)
		}
	}
	return nil
}
