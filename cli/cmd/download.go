package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/SheltonZhu/115driver/cli/internal/output"
	"github.com/SheltonZhu/115driver/cli/internal/resolver"
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/spf13/cobra"
)

const defaultDownloadTimeout = 2 * time.Hour

var downloadTimeout = defaultDownloadTimeout

var downloadCmd = &cobra.Command{
	Use:   "download <remote_path> <local_path>",
	Short: "Download a file from remote to a local directory or file path",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		localTarget := args[1]

		fileID, _, err := resolver.ResolvePath(client, remotePath)
		if err != nil {
			return &exitError{code: output.ExitNotFound, msg: err.Error()}
		}

		fileInfo, err := client.GetFile(fileID)
		if err != nil {
			return &exitError{code: output.ExitError, msg: err.Error()}
		}
		if fileInfo.IsDirectory {
			return &exitError{code: output.ExitArgs, msg: "Cannot download a directory."}
		}

		dlInfo, err := client.Download(fileInfo.PickCode)
		if err != nil {
			return &exitError{code: output.ExitError, msg: fmt.Sprintf("Failed to get download URL: %v", err)}
		}

		localPath := resolver.ResolveLocalDownloadPath(localTarget, dlInfo.FileName)

		if !jsonOutput {
			fmt.Printf("Downloading %s (%s)...\n", dlInfo.FileName, output.FormatFileSize(int64(dlInfo.FileSize)))
		}

		if err := downloadFile(dlInfo, localPath); err != nil {
			return &exitError{code: output.ExitError, msg: fmt.Sprintf("Download failed: %v", err)}
		}

		printer.PrintSuccess(map[string]interface{}{
			"remote_path": remotePath,
			"local_path":  localPath,
			"size":        int64(dlInfo.FileSize),
		})
		if !jsonOutput {
			fmt.Printf("Download complete: %s\n", localPath)
		}
		return nil
	},
}

func downloadFile(dlInfo *driver.DownloadInfo, localPath string) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}

	req, err := http.NewRequest("GET", dlInfo.Url.Url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	for k, vals := range dlInfo.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	resp, err := newDownloadHTTPClient(downloadTimeout).Do(req)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	return saveDownloadResponse(localPath, resp, 0)
}

func resolveDownloadTargetPath(localTarget, fileName string) string {
	return resolver.ResolveLocalDownloadPath(localTarget, fileName)
}

func newDownloadHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func saveDownloadResponse(localPath string, resp *http.Response, maxBytes int64) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(localPath), "."+filepath.Base(localPath)+".*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	var copyErr error
	if maxBytes > 0 {
		limited := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
		written, err := io.Copy(tmp, limited)
		if err != nil {
			copyErr = err
		} else if written > maxBytes || limited.N == 0 {
			copyErr = fmt.Errorf("download exceeds limit of %d bytes", maxBytes)
		}
	} else {
		_, copyErr = io.Copy(tmp, resp.Body)
	}
	if copyErr != nil {
		tmp.Close()
		return copyErr
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, localPath)
}

func init() {
	downloadCmd.Flags().DurationVar(&downloadTimeout, "timeout", defaultDownloadTimeout, "Download timeout, use 0 to disable")
	rootCmd.AddCommand(downloadCmd)
}
