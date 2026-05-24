package cmd

import (
	"github.com/SheltonZhu/115driver/cli/internal/output"
	"github.com/SheltonZhu/115driver/cli/internal/resolver"
	"github.com/spf13/cobra"
)

var lsLong bool
var lsOffset int64
var lsLimit int64

const (
	defaultLSLimit int64 = 100
	maxLSLimit     int64 = 500
)

var lsCmd = &cobra.Command{
	Use:   "ls [remote_path]",
	Short: "List directory contents",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := "/"
		if len(args) > 0 {
			remotePath = args[0]
		}

		dirID, err := resolver.ResolveDir(client, remotePath)
		if err != nil {
			return &exitError{code: output.ExitNotFound, msg: err.Error()}
		}

		offset, limit := normalizeLSPage(lsOffset, lsLimit)
		files, err := client.ListPage(dirID, offset, limit)
		if err != nil {
			return &exitError{code: output.ExitError, msg: err.Error()}
		}

		jsonFiles := make([]output.JSONFile, 0, len(*files))
		for _, f := range *files {
			jsonFiles = append(jsonFiles, output.FileToJSON(&f))
		}

		if lsLong {
			printer.PrintFileTable(remotePath, jsonFiles)
		} else {
			printer.PrintFileList(remotePath, jsonFiles)
		}
		return nil
	},
}

func init() {
	lsCmd.Flags().BoolVarP(&lsLong, "long", "l", false, "Show detailed listing")
	lsCmd.Flags().Int64Var(&lsOffset, "offset", 0, "Offset for paginated listing")
	lsCmd.Flags().Int64Var(&lsLimit, "limit", defaultLSLimit, "Max items to list")
	rootCmd.AddCommand(lsCmd)
}

func normalizeLSPage(offset, limit int64) (int64, int64) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultLSLimit
	}
	if limit > maxLSLimit {
		limit = maxLSLimit
	}
	return offset, limit
}
