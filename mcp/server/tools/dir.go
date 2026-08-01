package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DirTools holds directory-related MCP tools
type DirTools struct {
	client *driver.Pan115Client
}

const (
	maxDirectoryListLimit int64 = 500
)

// NewDirTools creates a new DirTools instance
func NewDirTools(client *driver.Pan115Client) *DirTools {
	return &DirTools{
		client: client,
	}
}

// ListDirectoryArgs defines arguments for list directory tool
type ListDirectoryArgs struct {
	DirID  string `json:"dir_id" jsonschema:"directory ID to list, default is root directory: 0"`
	Offset int64  `json:"offset,omitempty" jsonschema:"offset for pagination, default is 0"`
	Limit  int64  `json:"limit,omitempty" jsonschema:"number of items to return, default is all items"`
}

// RegisterTools registers directory-related tools with the MCP server
func (dt *DirTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listDirectory",
		Description: "List files and directories in a specific directory",
	}, dt.listDirectory)
}

func (dt *DirTools) listDirectory(ctx context.Context, req *mcp.CallToolRequest, args ListDirectoryArgs) (*mcp.CallToolResult, any, error) {
	var (
		files *[]driver.File
		err   error
	)

	offset, limit := normalizeDirectoryListPagination(args.Offset, args.Limit)
	if limit > 0 {
		files, err = dt.client.ListPage(args.DirID, offset, limit)
	} else {
		files, err = dt.client.List(args.DirID)
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list directory: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	// Serialize result to JSON
	resultJSON, err := json.Marshal(files)
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

func normalizeDirectoryListPagination(offset, limit int64) (int64, int64) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		return 0, 0
	}
	if limit > maxDirectoryListLimit {
		limit = maxDirectoryListLimit
	}
	return offset, limit
}
