package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RecycleTools holds recycle bin-related MCP tools
type RecycleTools struct {
	client           *driver.Pan115Client
	allowDestructive bool
}

type RecycleToolsOption func(*RecycleTools)

func WithRecycleDestructiveTools(allow bool) RecycleToolsOption {
	return func(rt *RecycleTools) {
		rt.allowDestructive = allow
	}
}

// NewRecycleTools creates a new RecycleTools instance
func NewRecycleTools(client *driver.Pan115Client, opts ...RecycleToolsOption) *RecycleTools {
	rt := &RecycleTools{
		client: client,
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

// ListRecycleArgs defines arguments for listing recycle bin items
type ListRecycleArgs struct {
	Offset int `json:"offset" jsonschema:"offset for pagination, default is 0"`
	Limit  int `json:"limit" jsonschema:"number of items to return, default is 40, maximum is 100"`
}

const (
	defaultRecycleLimit = 40
	maxRecycleLimit     = 100
)

// RevertRecycleArgs defines arguments for reverting recycle bin items
type RevertRecycleArgs struct {
	ItemIDs []string `json:"item_ids" jsonschema:"IDs of items to revert"`
}

// CleanRecycleArgs defines arguments for cleaning recycle bin items
type CleanRecycleArgs struct {
	Password string   `json:"password" jsonschema:"password for cleaning recycle bin"`
	ItemIDs  []string `json:"item_ids" jsonschema:"IDs of items to clean"`
}

// RegisterTools registers recycle bin-related tools with the MCP server
func (rt *RecycleTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listRecycleBin",
		Description: "List items in the recycle bin",
	}, rt.listRecycleBin)

	if rt.allowDestructive {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "revertRecycleBin",
			Description: "Revert items from the recycle bin",
		}, rt.revertRecycleBin)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "cleanRecycleBin",
			Description: "Clean items from the recycle bin",
		}, rt.cleanRecycleBin)
	}
}

func (rt *RecycleTools) listRecycleBin(ctx context.Context, req *mcp.CallToolRequest, args ListRecycleArgs) (*mcp.CallToolResult, any, error) {
	offset, limit := normalizeRecyclePagination(args.Offset, args.Limit)

	items, err := rt.client.ListRecycleBin(offset, limit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list recycle bin: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	resultJSON, err := json.Marshal(items)
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

func normalizeRecyclePagination(offset, limit int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultRecycleLimit
	}
	if limit > maxRecycleLimit {
		limit = maxRecycleLimit
	}
	return offset, limit
}

func (rt *RecycleTools) revertRecycleBin(ctx context.Context, req *mcp.CallToolRequest, args RevertRecycleArgs) (*mcp.CallToolResult, any, error) {
	if len(args.ItemIDs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No item IDs provided",
				},
			},
			IsError: true,
		}, nil, nil
	}

	err := rt.client.RevertRecycleBin(args.ItemIDs...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to revert recycle bin items: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Items reverted successfully",
			},
		},
	}, nil, nil
}

func (rt *RecycleTools) cleanRecycleBin(ctx context.Context, req *mcp.CallToolRequest, args CleanRecycleArgs) (*mcp.CallToolResult, any, error) {
	if len(args.ItemIDs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No item IDs provided",
				},
			},
			IsError: true,
		}, nil, nil
	}

	err := rt.client.CleanRecycleBin(args.Password, args.ItemIDs...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to clean recycle bin items: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Items cleaned successfully",
			},
		},
	}, nil, nil
}
