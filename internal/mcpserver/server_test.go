package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	internaldb "github.com/graditya/oxaudit/internal/db"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	d, err := internaldb.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return New(d, "test-run-001")
}

func serverRoundTrip(t *testing.T, srv *Server, req map[string]interface{}) map[string]interface{} {
	t.Helper()
	b, _ := json.Marshal(req)
	r := bytes.NewReader(append(b, '\n'))
	var w bytes.Buffer
	if err := srv.Serve(context.Background(), r, &w); err != nil {
		t.Fatalf("serve: %v", err)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, w.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	srv := newTestServer(t)
	resp := serverRoundTrip(t, srv, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]interface{}{},
	})
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result, got: %v", resp)
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocolVersion: %v", result["protocolVersion"])
	}
}

func TestToolsList(t *testing.T) {
	srv := newTestServer(t)
	resp := serverRoundTrip(t, srv, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	})
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result, got: %v", resp)
	}
	tools, ok := result["tools"].([]interface{})
	if !ok || len(tools) != 6 {
		t.Errorf("expected 6 tools, got: %v", result["tools"])
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := newTestServer(t)
	resp := serverRoundTrip(t, srv, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "unknown/method",
	})
	if resp["error"] == nil {
		t.Errorf("expected error for unknown method, got: %v", resp)
	}
}

func TestToolsCallMissingArguments(t *testing.T) {
	srv := newTestServer(t)
	// tools/call with no "arguments" key — valid for tools like get_summary
	resp := serverRoundTrip(t, srv, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "tools/call",
		"params":  map[string]interface{}{"name": "get_summary"},
	})
	if resp["error"] != nil {
		t.Errorf("expected no rpc error for missing arguments, got: %v", resp["error"])
	}
}

func TestNotificationIgnored(t *testing.T) {
	srv := newTestServer(t)
	// Notifications have no "id" — server must not respond
	notif, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	req, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	})
	input := append(notif, '\n')
	input = append(input, req...)
	input = append(input, '\n')

	var w bytes.Buffer
	srv.Serve(context.Background(), bytes.NewReader(input), &w)

	// Should have exactly one JSON object in output (the tools/list response)
	dec := json.NewDecoder(&w)
	var count int
	for dec.More() {
		var v interface{}
		dec.Decode(&v)
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 response, got %d", count)
	}
}
