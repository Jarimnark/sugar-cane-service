// Package mcp is a minimal Model Context Protocol server over the
// "Streamable HTTP" transport, stateless and JSON-only (no SSE streams).
//
// Every request is one JSON-RPC 2.0 message in a POST body. Requests get a
// JSON response; notifications get 202 Accepted with no body. This is enough
// for MCP clients such as Claude Code to list and call tools.
//
// The package is standard-library only on purpose: tools are plugged in from
// the outside (see the hopbu subpackage) so this layer stays easy to test.
package mcp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// LatestProtocolVersion is returned when the client asks for a version this
// server does not know.
const LatestProtocolVersion = "2025-06-18"

var supportedVersions = map[string]bool{
	"2025-06-18": true,
	"2025-03-26": true,
	"2024-11-05": true,
}

const maxBodyBytes = 1 << 20 // 1 MB per JSON-RPC message

// JSON-RPC 2.0 error codes.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
)

// Tool is one callable tool. InputSchema is a JSON Schema object.
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
	// Handler receives the raw "arguments" object. A returned error is shown
	// to the model as a tool error (isError=true), not as a protocol error.
	Handler func(ctx context.Context, args json.RawMessage) (any, error)
}

// Server serves MCP over HTTP.
type Server struct {
	name         string
	version      string
	instructions string
	token        string
	tools        []Tool
	byName       map[string]Tool
}

// NewServer builds a server. token is the shared bearer token every request
// must carry in "Authorization: Bearer <token>"; it must not be empty.
func NewServer(name, version, instructions, token string, tools []Tool) (*Server, error) {
	if token == "" {
		return nil, errors.New("mcp: token must not be empty")
	}
	s := &Server{
		name:         name,
		version:      version,
		instructions: instructions,
		token:        token,
		tools:        tools,
		byName:       make(map[string]Tool, len(tools)),
	}
	for _, t := range tools {
		if _, dup := s.byName[t.Name]; dup {
			return nil, errors.New("mcp: duplicate tool " + t.Name)
		}
		s.byName[t.Name] = t
	}
	return s, nil
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="hopbu-mcp"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodPost:
	case http.MethodDelete:
		// Stateless server: there is no session to end.
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		// GET would open an SSE stream; this server does not offer one.
		w.Header().Set("Allow", "POST, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req rpcRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, rpcResponse{JSONRPC: "2.0", ID: json.RawMessage("null"),
			Error: &rpcError{Code: codeParseError, Message: "parse error"}})
		return
	}
	if req.JSONRPC != "2.0" || req.Method == "" {
		writeJSON(w, rpcResponse{JSONRPC: "2.0", ID: idOrNull(req.ID),
			Error: &rpcError{Code: codeInvalidRequest, Message: "invalid request"}})
		return
	}

	// Notifications (no id) and client responses need no reply.
	if len(req.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	result, rpcErr := s.dispatch(r.Context(), req)
	resp := rpcResponse{JSONRPC: "2.0", ID: req.ID}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}
	writeJSON(w, resp)
}

func (s *Server) authorized(r *http.Request) bool {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return false
	}
	got := strings.TrimSpace(h[len(prefix):])
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.token)) == 1
}

func (s *Server) dispatch(ctx context.Context, req rpcRequest) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		version := p.ProtocolVersion
		if !supportedVersions[version] {
			version = LatestProtocolVersion
		}
		res := map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": s.name, "version": s.version},
		}
		if s.instructions != "" {
			res["instructions"] = s.instructions
		}
		return res, nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		list := make([]map[string]any, 0, len(s.tools))
		for _, t := range s.tools {
			schema := t.InputSchema
			if schema == nil {
				schema = map[string]any{"type": "object", "properties": map[string]any{}}
			}
			list = append(list, map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": schema,
			})
		}
		return map[string]any{"tools": list}, nil

	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil || p.Name == "" {
			return nil, &rpcError{Code: codeInvalidParams, Message: "invalid params: name is required"}
		}
		tool, ok := s.byName[p.Name]
		if !ok {
			return nil, &rpcError{Code: codeInvalidParams, Message: "unknown tool: " + p.Name}
		}
		args := p.Arguments
		if len(args) == 0 || string(args) == "null" {
			args = json.RawMessage("{}")
		}
		out, err := tool.Handler(ctx, args)
		if err != nil {
			slog.WarnContext(ctx, "mcp tool error", "tool", p.Name, "error", err)
			return toolResult(err.Error(), true), nil
		}
		slog.InfoContext(ctx, "mcp tool call", "tool", p.Name)
		text, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return toolResult("could not encode result", true), nil
		}
		return toolResult(string(text), false), nil

	default:
		return nil, &rpcError{Code: codeMethodNotFound, Message: "method not found: " + req.Method}
	}
}

func toolResult(text string, isError bool) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": isError,
	}
}

func idOrNull(id json.RawMessage) json.RawMessage {
	if len(id) == 0 {
		return json.RawMessage("null")
	}
	return id
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}
