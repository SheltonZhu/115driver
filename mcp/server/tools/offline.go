package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OfflineTools holds offline-related MCP tools
type OfflineTools struct {
	client           *driver.Pan115Client
	defaultSaveDir   string
	allowDestructive bool
}

type OfflineToolsOption func(*OfflineTools)

// WithOfflineDefaultSaveDir sets the default offline download directory name.
func WithOfflineDefaultSaveDir(dir string) OfflineToolsOption {
	return func(ot *OfflineTools) {
		ot.defaultSaveDir = dir
	}
}

// WithOfflineDestructiveTools controls whether destructive offline tools are registered.
func WithOfflineDestructiveTools(allow bool) OfflineToolsOption {
	return func(ot *OfflineTools) {
		ot.allowDestructive = allow
	}
}

// NewOfflineTools creates a new OfflineTools instance
func NewOfflineTools(client *driver.Pan115Client, opts ...OfflineToolsOption) *OfflineTools {
	ot := &OfflineTools{
		client: client,
	}
	for _, opt := range opts {
		opt(ot)
	}
	return ot
}

// ListOfflineTaskArgs defines arguments for listing offline tasks
type ListOfflineTaskArgs struct {
	Page int64 `json:"page" jsonschema:"page number for pagination, default is 1"`
}

// AddOfflineTaskURIsArgs defines arguments for adding offline tasks
type AddOfflineTaskURIsArgs struct {
	URIs      []string `json:"uris" jsonschema:"download URIs, supports http, ed2k, magnet"`
	SaveDirID string   `json:"save_dir_id,omitempty" jsonschema:"directory ID to save downloaded files, leave empty to use config default"`
}

// DeleteOfflineTasksArgs defines arguments for deleting offline tasks
type DeleteOfflineTasksArgs struct {
	Hashes      []string `json:"hashes" jsonschema:"task hashes to delete"`
	DeleteFiles bool     `json:"delete_files" jsonschema:"whether to delete associated files, default is false"`
}

// ClearOfflineTasksArgs defines arguments for clearing offline tasks
type ClearOfflineTasksArgs struct {
	ClearFlag int64 `json:"clear_flag" jsonschema:"clear flag, 0: clear completed tasks, 1: clear all tasks"`
}

// RegisterTools registers offline-related tools with the MCP server
func (ot *OfflineTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listOfflineTasks",
		Description: "List offline download tasks",
	}, ot.listOfflineTasks)

	if ot.allowDestructive {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "addOfflineTaskURIs",
			Description: "Add offline tasks by download URIs, supports http, ed2k, magnet",
		}, ot.addOfflineTaskURIs)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "deleteOfflineTasks",
			Description: "Delete offline tasks",
		}, ot.deleteOfflineTasks)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "clearOfflineTasks",
			Description: "Clear offline tasks",
		}, ot.clearOfflineTasks)
	}
}

func (ot *OfflineTools) listOfflineTasks(ctx context.Context, req *mcp.CallToolRequest, args ListOfflineTaskArgs) (*mcp.CallToolResult, any, error) {
	page := args.Page
	if page <= 0 {
		page = 1
	}

	result, err := ot.client.ListOfflineTask(page)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list offline tasks: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	// Convert tasks to a serializable format
	tasks := make([]map[string]interface{}, len(result.Tasks))
	for i, task := range result.Tasks {
		tasks[i] = map[string]interface{}{
			"info_hash":      task.InfoHash,
			"name":           task.Name,
			"size":           task.Size,
			"url":            task.Url,
			"add_time":       task.AddTime,
			"peers":          task.Peers,
			"rate_download":  task.RateDownload,
			"status":         task.Status,
			"status_text":    task.GetStatus(),
			"percent":        task.Percent,
			"update_time":    task.UpdateTime,
			"left_time":      task.LeftTime,
			"file_id":        task.FileId,
			"delete_file_id": task.DelFileId,
			"dir_id":         task.DirId,
			"move":           task.Move,
		}
	}

	response := map[string]interface{}{
		"total":      result.Total,
		"count":      result.Count,
		"page_row":   result.PageRow,
		"page_count": result.PageCount,
		"page":       result.Page,
		"quota":      result.Quota,
		"tasks":      tasks,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize offline tasks: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseJSON),
			},
		},
	}, nil, nil
}

func (ot *OfflineTools) addOfflineTaskURIs(ctx context.Context, req *mcp.CallToolRequest, args AddOfflineTaskURIsArgs) (*mcp.CallToolResult, any, error) {
	if len(args.URIs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No URIs provided",
				},
			},
			IsError: true,
		}, nil, nil
	}

	saveDirID := args.SaveDirID
	if saveDirID == "" && ot.defaultSaveDir != "" {
		// Resolve default save directory name to ID
		resp, err := ot.client.DirName2CID(ot.defaultSaveDir)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Default save directory not found (from config default_offline_save_dir): %s", ot.defaultSaveDir),
					},
				},
				IsError: true,
			}, nil, nil
		}
		if string(resp.CategoryID) == "0" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Default save directory not found (from config default_offline_save_dir): %s", ot.defaultSaveDir),
					},
				},
				IsError: true,
			}, nil, nil
		}
		saveDirID = string(resp.CategoryID)
	}
	if saveDirID == "" {
		saveDirID = "0"
	}

	hashes, err := ot.client.AddOfflineTaskURIs(args.URIs, saveDirID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to add offline tasks: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	response := map[string]interface{}{
		"hashes": hashes,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to serialize response: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseJSON),
			},
		},
	}, nil, nil
}

func (ot *OfflineTools) deleteOfflineTasks(ctx context.Context, req *mcp.CallToolRequest, args DeleteOfflineTasksArgs) (*mcp.CallToolResult, any, error) {
	if len(args.Hashes) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No task hashes provided",
				},
			},
			IsError: true,
		}, nil, nil
	}

	err := ot.client.DeleteOfflineTasks(args.Hashes, args.DeleteFiles)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to delete offline tasks: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully deleted offline tasks",
			},
		},
	}, nil, nil
}

func (ot *OfflineTools) clearOfflineTasks(ctx context.Context, req *mcp.CallToolRequest, args ClearOfflineTasksArgs) (*mcp.CallToolResult, any, error) {
	err := ot.client.ClearOfflineTasks(args.ClearFlag)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to clear offline tasks: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully cleared offline tasks",
			},
		},
	}, nil, nil
}
