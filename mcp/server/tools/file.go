package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FileTools holds file-related MCP tools
type FileTools struct {
	client            *driver.Pan115Client
	localRoot         string
	downloadTimeout   time.Duration
	urlUploadMaxBytes int64
	downloadMaxBytes  int64
	allowDestructive  bool
}

type FileToolsOption func(*FileTools)

func WithLocalRoot(root string) FileToolsOption {
	return func(ft *FileTools) {
		ft.localRoot = root
	}
}

func WithDownloadTimeout(timeout time.Duration) FileToolsOption {
	return func(ft *FileTools) {
		ft.downloadTimeout = timeout
	}
}

func WithURLUploadMaxBytes(maxBytes int64) FileToolsOption {
	return func(ft *FileTools) {
		ft.urlUploadMaxBytes = maxBytes
	}
}

func WithDownloadMaxBytes(maxBytes int64) FileToolsOption {
	return func(ft *FileTools) {
		ft.downloadMaxBytes = maxBytes
	}
}

func WithDestructiveTools(allow bool) FileToolsOption {
	return func(ft *FileTools) {
		ft.allowDestructive = allow
	}
}

// NewFileTools creates a new FileTools instance
func NewFileTools(client *driver.Pan115Client, opts ...FileToolsOption) *FileTools {
	ft := &FileTools{
		client:            client,
		downloadTimeout:   defaultMCPDownloadTimeout,
		urlUploadMaxBytes: defaultMCPURLUploadMaxBytes,
		downloadMaxBytes:  defaultMCPDownloadMaxBytes,
	}
	for _, opt := range opts {
		opt(ft)
	}
	return ft
}

const (
	defaultMCPURLUploadMaxBytes int64 = 2 << 30 // 2 GiB
	defaultMCPDownloadMaxBytes  int64 = 0       // unlimited
	defaultMCPDownloadTimeout         = 2 * time.Hour
)

var (
	errUnexpectedHTTPStatus = errors.New("unexpected HTTP status")
	errResponseTooLarge     = errors.New("response too large")
	errInvalidSizeLimit     = errors.New("invalid size limit")
)

// MkdirArgs defines arguments for mkdir tool
type MkdirArgs struct {
	ParentID string `json:"parent_id" jsonschema:"parent directory ID"`
	Name     string `json:"name" jsonschema:"name of the new directory"`
}

// DeleteArgs defines arguments for delete tool
type DeleteArgs struct {
	FileIDs []string `json:"file_ids" jsonschema:"IDs of files or directories to delete"`
}

// RenameArgs defines arguments for rename tool
type RenameArgs struct {
	FileID  string `json:"file_id" jsonschema:"ID of file or directory to rename"`
	NewName string `json:"new_name" jsonschema:"new name for the file or directory"`
}

// BatchRenameArgs defines arguments for the batch rename tool.
type BatchRenameArgs struct {
	Renames []driver.RenameItem `json:"renames" jsonschema:"files or directories to rename"`
}

// MoveArgs defines arguments for move tool
type MoveArgs struct {
	DirID   string   `json:"dir_id" jsonschema:"target directory ID"`
	FileIDs []string `json:"file_ids" jsonschema:"IDs of files or directories to move"`
}

// CopyArgs defines arguments for copy tool
type CopyArgs struct {
	DirID   string   `json:"dir_id" jsonschema:"target directory ID"`
	FileIDs []string `json:"file_ids" jsonschema:"IDs of files or directories to copy"`
}

// StatArgs defines arguments for stat tool
type StatArgs struct {
	FileID string `json:"file_id" jsonschema:"ID of file or directory to get info"`
}

// UploadFromURLArgs defines arguments for uploading from URL
type UploadFromURLArgs struct {
	URL      string `json:"url" jsonschema:"URL of the file to download and upload"`
	DirID    string `json:"dir_id" jsonschema:"target directory ID in 115 cloud for saving the file"`
	FileName string `json:"file_name,omitempty" jsonschema:"optional filename for the uploaded file, defaults to original filename"`
}

// UploadFromLocalArgs defines arguments for uploading from local file
type UploadFromLocalArgs struct {
	LocalPath string `json:"local_path" jsonschema:"absolute path to the local file to upload"`
	DirID     string `json:"dir_id" jsonschema:"target directory ID in 115 cloud"`
	FileName  string `json:"file_name,omitempty" jsonschema:"optional filename for the uploaded file, defaults to original filename"`
}

// DownloadFileArgs defines arguments for downloading a file
type DownloadFileArgs struct {
	PickCode  string `json:"pick_code" jsonschema:"pick code of the file to download"`
	LocalPath string `json:"local_path" jsonschema:"local path where the downloaded file will be saved"`
	UserAgent string `json:"user_agent,omitempty" jsonschema:"optional user agent for the download request, uses 115 browser UA if not specified"`
}

// GetDownloadInfoArgs defines arguments for getting download information
type GetDownloadInfoArgs struct {
	PickCode  string `json:"pick_code" jsonschema:"pick code of the file to get info for"`
	UserAgent string `json:"user_agent,omitempty" jsonschema:"optional user agent for the download request, uses 115 browser UA if not specified"`
}

// GetDownloadInfoResult defines the result for getting download information
type GetDownloadInfoResult struct {
	URL      string `json:"url" jsonschema:"download URL"`
	FileName string `json:"file_name" jsonschema:"file name"`
	Size     int64  `json:"size" jsonschema:"file size in bytes"`
}

// DownloadFileResult defines the result for downloading a file
type DownloadFileResult struct {
	Message string `json:"message" jsonschema:"result message"`
}

// RegisterTools registers file-related tools with the MCP server
func (ft *FileTools) RegisterTools(server *mcp.Server) {
	if ft.allowDestructive {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "mkdir",
			Description: "Create a new directory",
		}, ft.mkdir)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "delete",
			Description: "Delete files or directories",
		}, ft.delete)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "rename",
			Description: "Rename a file or directory",
		}, ft.rename)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "batch_rename",
			Description: "Rename multiple files or directories in one request",
		}, ft.batchRename)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "move",
			Description: "Move files or directories to another directory",
		}, ft.move)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "copy",
			Description: "Copy files or directories to another directory",
		}, ft.copy)
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "stat",
		Description: "Get detailed information about a file or directory",
	}, ft.stat)

	if ft.allowDestructive {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "upload_from_url",
			Description: "Upload a file to 115 cloud storage from a URL",
		}, ft.uploadFromURL)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "upload_from_local",
			Description: "Upload a local file to 115 cloud storage",
		}, ft.uploadFromLocal)
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "download_file",
		Description: "Download a file from 115 cloud storage to local path",
	}, ft.downloadFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_download_info",
		Description: "Get download information for a file including URL, file name, and size",
	}, ft.getDownloadInfo)
}

func (ft *FileTools) mkdir(ctx context.Context, req *mcp.CallToolRequest, args MkdirArgs) (*mcp.CallToolResult, any, error) {
	dirID, err := ft.client.Mkdir(args.ParentID, args.Name)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to create directory: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result := map[string]string{
		"directory_id": dirID,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize result: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func (ft *FileTools) delete(ctx context.Context, req *mcp.CallToolRequest, args DeleteArgs) (*mcp.CallToolResult, any, error) {
	if len(args.FileIDs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No file IDs provided",
				},
			},
			IsError: true,
		}, nil, nil
	}

	err := ft.client.Delete(args.FileIDs...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to delete files: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Files deleted successfully",
			},
		},
	}, nil, nil
}

func (ft *FileTools) rename(ctx context.Context, req *mcp.CallToolRequest, args RenameArgs) (*mcp.CallToolResult, any, error) {
	return ft.renameItems(driver.RenameItem{FileID: args.FileID, NewName: args.NewName})
}

func (ft *FileTools) batchRename(ctx context.Context, req *mcp.CallToolRequest, args BatchRenameArgs) (*mcp.CallToolResult, any, error) {
	return ft.renameItems(args.Renames...)
}

func (ft *FileTools) renameItems(items ...driver.RenameItem) (*mcp.CallToolResult, any, error) {
	if len(items) == 0 {
		return toolError("No renames provided"), nil, nil
	}

	err := ft.client.BatchRename(items...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to rename files: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Renamed %d file(s) successfully", len(items)),
			},
		},
	}, nil, nil
}

func (ft *FileTools) move(ctx context.Context, req *mcp.CallToolRequest, args MoveArgs) (*mcp.CallToolResult, any, error) {
	if len(args.FileIDs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No file IDs provided",
				},
			},
			IsError: true,
		}, nil, nil
	}

	err := ft.client.Move(args.DirID, args.FileIDs...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to move files: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Files moved successfully",
			},
		},
	}, nil, nil
}

func (ft *FileTools) copy(ctx context.Context, req *mcp.CallToolRequest, args CopyArgs) (*mcp.CallToolResult, any, error) {
	if len(args.FileIDs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No file IDs provided",
				},
			},
			IsError: true,
		}, nil, nil
	}

	err := ft.client.Copy(args.DirID, args.FileIDs...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to copy files: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Files copied successfully",
			},
		},
	}, nil, nil
}

func (ft *FileTools) stat(ctx context.Context, req *mcp.CallToolRequest, args StatArgs) (*mcp.CallToolResult, any, error) {
	info, err := ft.client.Stat(args.FileID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get file info: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result := map[string]interface{}{
		"name":         info.Name,
		"pick_code":    info.PickCode,
		"sha1":         info.Sha1,
		"is_directory": info.IsDirectory,
		"file_count":   info.FileCount,
		"dir_count":    info.DirCount,
		"create_time":  info.CreateTime,
		"update_time":  info.UpdateTime,
		"parents":      info.Parents,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize result: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func (ft *FileTools) uploadFromURL(ctx context.Context, req *mcp.CallToolRequest, args UploadFromURLArgs) (*mcp.CallToolResult, any, error) {
	downloadURL, err := validateUploadURL(args.URL)
	if err != nil {
		return toolError(fmt.Sprintf("Invalid URL: %v", err)), nil, nil
	}

	// Download the file from the URL
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL.String(), nil)
	if err != nil {
		return toolError(fmt.Sprintf("Failed to create download request: %v", err)), nil, nil
	}
	resp, err := newMCPHTTPClient(ft.downloadTimeout).Do(httpReq)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to download file from URL: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}
	defer resp.Body.Close()

	// If fileName is empty, try to extract it from the URL
	fileName := args.FileName
	if fileName == "" {
		fileName = filepath.Base(downloadURL.Path)
		if fileName == "" || fileName == "." || fileName == "/" {
			fileName = "downloaded_file"
		}
	}

	// Create a temporary file to store the downloaded content
	tempFile, err := os.CreateTemp("", "115_mcp_upload_*")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to create temporary file: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file afterwards
	defer tempFile.Close()

	// Copy the response body to the temporary file
	err = copyHTTPResponse(tempFile, resp, ft.urlUploadMaxBytes)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to save downloaded content to temporary file: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	// Get the file size
	fileInfo, err := tempFile.Stat()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get file info: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}
	fileSize := fileInfo.Size()

	// Seek back to the beginning of the file
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to seek to beginning of file: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	// Upload the downloaded content to 115 using the existing method
	err = ft.client.RapidUploadOrByOSS(args.DirID, fileName, fileSize, tempFile)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to upload file to 115: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result := map[string]string{
		"message": "File uploaded successfully from URL",
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize result: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func (ft *FileTools) uploadFromLocal(ctx context.Context, req *mcp.CallToolRequest, args UploadFromLocalArgs) (*mcp.CallToolResult, any, error) {
	localPath, err := validateLocalPath(ft.localRoot, args.LocalPath, true)
	if err != nil {
		return toolError(fmt.Sprintf("Local file access denied: %v", err)), nil, nil
	}

	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to open local file: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}
	defer file.Close()

	// Get file info to determine file size
	fileInfo, err := file.Stat()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get file info: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	// If fileName is empty, use the basename of the local file
	fileName := args.FileName
	if fileName == "" {
		fileName = fileInfo.Name()
	}

	// Get file size
	fileSize := fileInfo.Size()

	// Seek to the beginning of the file
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to seek to beginning of file: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	// Upload the file using the existing method
	err = ft.client.RapidUploadOrByOSS(args.DirID, fileName, fileSize, file)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to upload file to 115: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result := map[string]string{
		"message": "Local file uploaded successfully",
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize result: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func (ft *FileTools) downloadFile(ctx context.Context, req *mcp.CallToolRequest, args DownloadFileArgs) (*mcp.CallToolResult, any, error) {
	localPath, err := validateLocalPath(ft.localRoot, args.LocalPath, false)
	if err != nil {
		return toolError(fmt.Sprintf("Local file access denied: %v", err)), nil, nil
	}

	// Get download info with the specified User-Agent
	downloadInfo, err := ft.client.DownloadWithUA(args.PickCode, args.UserAgent)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get download info: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	// Perform the actual download using the same User-Agent
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadInfo.Url.Url, nil)
	if err != nil {
		return toolError(fmt.Sprintf("Failed to create download request: %v", err)), nil, nil
	}
	for k, vals := range downloadInfo.Header {
		for _, v := range vals {
			httpReq.Header.Add(k, v)
		}
	}

	resp, err := newMCPHTTPClient(ft.downloadTimeout).Do(httpReq)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to download file: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}
	defer resp.Body.Close()

	if err := saveHTTPResponseToFile(localPath, resp, ft.downloadMaxBytes); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to save file locally: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result := DownloadFileResult{
		Message: "File downloaded successfully",
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize result: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func (ft *FileTools) getDownloadInfo(ctx context.Context, req *mcp.CallToolRequest, args GetDownloadInfoArgs) (*mcp.CallToolResult, any, error) {
	// Get download info with the specified User-Agent
	downloadInfo, err := ft.client.DownloadWithUA(args.PickCode, args.UserAgent)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to get download info: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result := GetDownloadInfoResult{
		URL:      downloadInfo.Url.Url,
		FileName: downloadInfo.FileName,
		Size:     int64(downloadInfo.FileSize),
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize result: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func validateUploadURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
	host := parsed.Hostname()
	if host == "" {
		return nil, errors.New("missing host")
	}
	if isUnsafeHost(host) {
		return nil, fmt.Errorf("host %q is not allowed", host)
	}
	return parsed, nil
}

func isUnsafeHost(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return isUnsafeIP(ip)
	}
	return false
}

func validateResolvedIPs(host string, ips []net.IP) error {
	if len(ips) == 0 {
		return fmt.Errorf("host %q did not resolve to any IPs", host)
	}
	for _, ip := range ips {
		if ip == nil || isUnsafeIP(ip) {
			return fmt.Errorf("host %q resolved to unsafe IP %s", host, ip)
		}
	}
	return nil
}

func dialResolvedIPs(ctx context.Context, network, host, port string, ips []net.IP, dial func(context.Context, string, string) (net.Conn, error)) (net.Conn, error) {
	if err := validateResolvedIPs(host, ips); err != nil {
		return nil, err
	}

	var errs []error
	for _, ip := range ips {
		address := net.JoinHostPort(ip.String(), port)
		conn, err := dial(ctx, network, address)
		if err == nil {
			return conn, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", address, err))
	}
	return nil, fmt.Errorf("dial %q: %w", host, errors.Join(errs...))
}

func isUnsafeIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}

func validateLocalPath(root, target string, mustExist bool) (string, error) {
	if root == "" {
		return "", errors.New("local root is not configured")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", fmt.Errorf("resolve local root: %w", err)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	pathToCheck := absTarget
	if !mustExist && !pathExists(absTarget) {
		pathToCheck, err = nearestExistingPath(filepath.Dir(absTarget))
		if err != nil {
			return "", err
		}
	}
	realCheck, err := filepath.EvalSymlinks(pathToCheck)
	if err != nil {
		if mustExist || !os.IsNotExist(err) {
			return "", err
		}
		realCheck = pathToCheck
	}

	rel, err := filepath.Rel(realRoot, realCheck)
	if err != nil {
		return "", err
	}
	if rel == ".." || len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator) {
		return "", errors.New("path escapes local root")
	}
	return absTarget, nil
}

func nearestExistingPath(path string) (string, error) {
	for {
		if pathExists(path) {
			return path, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("no existing parent for %q", path)
		}
		path = parent
	}
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func copyHTTPResponse(dst io.Writer, resp *http.Response, maxBytes int64) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", errUnexpectedHTTPStatus, resp.StatusCode)
	}
	if maxBytes < 0 {
		return fmt.Errorf("%w: %d", errInvalidSizeLimit, maxBytes)
	}
	if maxBytes == 0 {
		_, err := io.Copy(dst, resp.Body)
		return err
	}
	limited := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
	written, err := io.Copy(dst, limited)
	if err != nil {
		return err
	}
	if written > maxBytes || limited.N == 0 {
		return fmt.Errorf("%w: limit is %d bytes", errResponseTooLarge, maxBytes)
	}
	return nil
}

func newMCPHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if _, err := validateUploadURL(req.URL.String()); err != nil {
				return err
			}
			return nil
		},
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(address)
				if err != nil {
					return nil, err
				}
				ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
				if err != nil {
					return nil, err
				}
				return dialResolvedIPs(ctx, network, host, port, ips, dialer.DialContext)
			},
		},
	}
}

func saveHTTPResponseToFile(path string, resp *http.Response, maxBytes int64) error {
	if maxBytes < 0 {
		return fmt.Errorf("%w: %d", errInvalidSizeLimit, maxBytes)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := copyHTTPResponse(tmp, resp, maxBytes); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func toolError(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
		IsError: true,
	}
}
