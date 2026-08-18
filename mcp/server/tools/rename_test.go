package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestBatchRenameRejectsEmptyInput(t *testing.T) {
	ft := NewFileTools(nil)
	result, _, err := ft.batchRename(context.Background(), nil, BatchRenameArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected an MCP tool error")
	}
}

func TestBatchRenameRegistration(t *testing.T) {
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "test"}, nil)
	NewFileTools(nil, WithDestructiveTools(true)).RegisterTools(server)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range result.Tools {
		if tool.Name == "batch_rename" {
			return
		}
	}
	t.Fatal("batch_rename tool was not registered")
}
