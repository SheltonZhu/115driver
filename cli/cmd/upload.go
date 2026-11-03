package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SheltonZhu/115driver/cli/internal/output"
	"github.com/SheltonZhu/115driver/cli/internal/resolver"
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <local_path> <remote_dir>",
	Short: "Upload a file to remote directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		localPath := args[0]
		remoteDir := args[1]

		dirID, err := resolver.ResolveDir(client, remoteDir)
		if err != nil {
			return &exitError{code: output.ExitNotFound, msg: fmt.Sprintf("Remote directory not found: %s", remoteDir)}
		}

		f, err := os.Open(localPath)
		if err != nil {
			return &exitError{code: output.ExitArgs, msg: fmt.Sprintf("Cannot open local file: %v", err)}
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			return &exitError{code: output.ExitError, msg: err.Error()}
		}

		if stat.IsDir() {
			return &exitError{code: output.ExitArgs, msg: "Directory upload is not supported. Upload individual files."}
		}

		fileName := filepath.Base(localPath)

		if !jsonOutput {
			fmt.Printf("Uploading %s (%s)...\n", fileName, output.FormatFileSize(stat.Size()))
		}

		if stat.Size() >= 5*driver.GB {
			err = client.RapidUploadOrByMultipart(dirID, fileName, stat.Size(), f)
		} else {
			err = client.RapidUploadOrByOSS(dirID, fileName, stat.Size(), f)
		}
		if err != nil {
			return &exitError{code: output.ExitError, msg: fmt.Sprintf("Upload failed: %v", err)}
		}

		printer.PrintSuccess(map[string]interface{}{
			"local_path": localPath,
			"remote_dir": remoteDir,
			"size":       stat.Size(),
		})
		if !jsonOutput {
			fmt.Printf("Upload complete: %s -> %s\n", fileName, remoteDir)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
