package mcpserver

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Server is the oxaudit MCP server.
type Server struct {
	db         *sql.DB
	mu         sync.RWMutex // protects auditRunID
	auditRunID string
	runMu      sync.Mutex // held while run_audit executes
}

// New creates a Server backed by the given database and initial audit run ID.
func New(db *sql.DB, auditRunID string) *Server {
	return &Server{db: db, auditRunID: auditRunID}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CallResult is the standard MCP tool result envelope.
type CallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content is a single content block in a CallResult.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Serve reads newline-delimited JSON-RPC requests from r and writes responses to w.
func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	enc := json.NewEncoder(w)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue // malformed line — skip
		}

		// Notifications have no "id" field — do not respond.
		if len(req.ID) == 0 {
			continue
		}

		result, rpcErr := s.dispatch(ctx, req)
		resp := rpcResponse{JSONRPC: "2.0", ID: req.ID}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("encoding response: %w", err)
		}
	}
	return scanner.Err()
}

func (s *Server) dispatch(ctx context.Context, req rpcRequest) (interface{}, *rpcError) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize()
	case "tools/list":
		return s.handleToolsList()
	case "tools/call":
		return s.handleToolsCall(ctx, req.Params)
	case "ping":
		return map[string]interface{}{}, nil
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found: " + req.Method}
	}
}

func (s *Server) handleInitialize() (interface{}, *rpcError) {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
		"serverInfo":      map[string]interface{}{"name": "oxaudit", "version": "0.1.0"},
	}, nil
}

func (s *Server) handleToolsList() (interface{}, *rpcError) {
	return map[string]interface{}{"tools": allTools()}, nil
}

type toolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, *rpcError) {
	var p toolsCallParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &rpcError{Code: -32602, Message: "invalid params"}
		}
	}
	if p.Arguments == nil {
		p.Arguments = map[string]interface{}{}
	}

	var result CallResult
	var rpcErr *rpcError

	switch p.Name {
	case "run_audit":
		result, rpcErr = s.handleRunAudit(ctx, p.Arguments)
	case "get_summary":
		result, rpcErr = s.handleGetSummary(ctx, p.Arguments)
	case "list_findings":
		result, rpcErr = s.handleListFindings(ctx, p.Arguments)
	case "get_finding":
		result, rpcErr = s.handleGetFinding(ctx, p.Arguments)
	case "get_cost_breakdown":
		result, rpcErr = s.handleGetCostBreakdown(ctx, p.Arguments)
	case "query_resources":
		result, rpcErr = s.handleQueryResources(ctx, p.Arguments)
	default:
		return nil, &rpcError{Code: -32602, Message: "unknown tool: " + p.Name}
	}

	if rpcErr != nil {
		return nil, rpcErr
	}
	return result, nil
}
