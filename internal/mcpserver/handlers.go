package mcpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

func textOK(v interface{}) (CallResult, *rpcError) {
	b, _ := json.MarshalIndent(v, "", "  ")
	return CallResult{Content: []Content{{Type: "text", Text: string(b)}}}, nil
}

func textErr(msg string) (CallResult, *rpcError) {
	return CallResult{Content: []Content{{Type: "text", Text: msg}}, IsError: true}, nil
}

func strArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intArg(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
}

func floatArg(args map[string]interface{}, key string, def float64) float64 {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return def
}

func (s *Server) handleGetSummary(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

func (s *Server) handleListFindings(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

func (s *Server) handleGetFinding(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

func (s *Server) handleGetCostBreakdown(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

func (s *Server) handleQueryResources(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

func (s *Server) handleRunAudit(_ context.Context, _ map[string]interface{}) (CallResult, *rpcError) {
	return textErr("not yet implemented")
}

// Ensure sql, fmt packages are used by later tasks — import them now to avoid churn.
var _ = sql.ErrNoRows
var _ = fmt.Sprintf
