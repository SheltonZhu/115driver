package server

import (
	"context"
	"log"
	"time"

	"github.com/SheltonZhu/115driver/mcp/server/tools"
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the 115driver MCP server
type Server struct {
	mcpServer         *mcp.Server
	client            *driver.Pan115Client
	localRoot         string
	downloadTimeout   time.Duration
	urlUploadMaxBytes int64
	downloadMaxBytes  int64
	allowDestructive  bool
}

// NewServer creates a new 115driver MCP server
func NewServer() *Server {
	return &Server{
		downloadTimeout:   2 * time.Hour,
		urlUploadMaxBytes: 2 << 30,
		downloadMaxBytes:  0,
		mcpServer: mcp.NewServer(&mcp.Implementation{
			Name:    "115driver-mcp-server",
			Version: "1.0.0",
		}, nil),
	}
}

// WithClient sets the 115 driver client for the server
func (s *Server) WithClient(client *driver.Pan115Client) *Server {
	s.client = client
	return s
}

// WithLocalRoot restricts local file tools to paths under root. Empty disables them.
func (s *Server) WithLocalRoot(root string) *Server {
	s.localRoot = root
	return s
}

// WithDownloadTimeout sets the total timeout for MCP HTTP transfers. Zero disables it.
func (s *Server) WithDownloadTimeout(timeout time.Duration) *Server {
	s.downloadTimeout = timeout
	return s
}

// WithTransferSizeLimits sets size limits for MCP HTTP transfers. Zero disables a limit.
func (s *Server) WithTransferSizeLimits(urlUploadMaxBytes, downloadMaxBytes int64) *Server {
	s.urlUploadMaxBytes = urlUploadMaxBytes
	s.downloadMaxBytes = downloadMaxBytes
	return s
}

// WithDestructiveTools controls MCP tools that mutate 115 cloud state.
func (s *Server) WithDestructiveTools(allow bool) *Server {
	s.allowDestructive = allow
	return s
}

// Start runs the MCP server
func (s *Server) Start(ctx context.Context) error {
	// Register all tools
	s.registerTools()

	// Run the server on the stdio transport
	if err := s.mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Printf("Server failed: %v", err)
		return err
	}
	return nil
}

// registerTools registers all available tools with the MCP server
func (s *Server) registerTools() {
	// Register account tools
	accountTools := tools.NewAccountTools(s.client)
	accountTools.RegisterTools(s.mcpServer)

	// Register directory tools
	dirTools := tools.NewDirTools(s.client)
	dirTools.RegisterTools(s.mcpServer)

	// Register file tools
	fileTools := tools.NewFileTools(
		s.client,
		tools.WithLocalRoot(s.localRoot),
		tools.WithDownloadTimeout(s.downloadTimeout),
		tools.WithURLUploadMaxBytes(s.urlUploadMaxBytes),
		tools.WithDownloadMaxBytes(s.downloadMaxBytes),
		tools.WithDestructiveTools(s.allowDestructive),
	)
	fileTools.RegisterTools(s.mcpServer)

	// Register recycle tools
	recycleTools := tools.NewRecycleTools(s.client, tools.WithRecycleDestructiveTools(s.allowDestructive))
	recycleTools.RegisterTools(s.mcpServer)

	// Register share tools
	shareTools := tools.NewShareTools(s.client)
	shareTools.RegisterTools(s.mcpServer)

	// Register search tools
	searchTools := tools.NewSearchTools(s.client)
	searchTools.RegisterTools(s.mcpServer)

	// Register offline tools
	offlineTools := tools.NewOfflineTools(s.client, tools.WithOfflineDestructiveTools(s.allowDestructive))
	offlineTools.RegisterTools(s.mcpServer)
}
