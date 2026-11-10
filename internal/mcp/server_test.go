package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testToken = "test-token-0123456789abcdef0123456789"

func newTestServer(t *testing.T) *Server {
	t.Helper()
	s, err := NewServer("hopbu", "1.0.0", "hello", testToken, []Tool{
		{
			Name:        "echo",
			Description: "echoes msg",
			Handler: func(_ context.Context, args json.RawMessage) (any, error) {
				var a struct {
					Msg string `json:"msg"`
				}
				_ = json.Unmarshal(args, &a)
				return map[string]string{"msg": a.Msg}, nil
			},
		},
		{
			Name: "fail",
			Handler: func(context.Context, json.RawMessage) (any, error) {
				return nil, errors.New("boom")
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func do(t *testing.T, s *Server, method, auth, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/mcp", strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	return rec
}

func decodeResp(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("bad json %q: %v", rec.Body.String(), err)
	}
	return m
}

func TestNewServerRejectsEmptyToken(t *testing.T) {
	if _, err := NewServer("x", "1", "", "", nil); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestAuth(t *testing.T) {
	s := newTestServer(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	for _, auth := range []string{"", "Bearer wrong", "Basic " + testToken, testToken} {
		if rec := do(t, s, http.MethodPost, auth, body); rec.Code != http.StatusUnauthorized {
			t.Errorf("auth %q: got %d, want 401", auth, rec.Code)
		}
	}
	if rec := do(t, s, http.MethodPost, "bearer "+testToken, body); rec.Code != http.StatusOK {
		t.Errorf("valid token: got %d", rec.Code)
	}
}

func TestInitialize(t *testing.T) {
	s := newTestServer(t)
	auth := "Bearer " + testToken

	m := decodeResp(t, do(t, s, http.MethodPost, auth,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`))
	res := m["result"].(map[string]any)
	if res["protocolVersion"] != "2025-03-26" {
		t.Errorf("version = %v, want echo of 2025-03-26", res["protocolVersion"])
	}
	if res["instructions"] != "hello" {
		t.Errorf("instructions = %v", res["instructions"])
	}
	if _, ok := res["capabilities"].(map[string]any)["tools"]; !ok {
		t.Error("tools capability missing")
	}

	m = decodeResp(t, do(t, s, http.MethodPost, auth,
		`{"jsonrpc":"2.0","id":"a","method":"initialize","params":{"protocolVersion":"1999-01-01"}}`))
	if v := m["result"].(map[string]any)["protocolVersion"]; v != LatestProtocolVersion {
		t.Errorf("unknown version answered with %v", v)
	}
	if m["id"] != "a" {
		t.Errorf("id not echoed: %v", m["id"])
	}
}

func TestNotificationGets202(t *testing.T) {
	s := newTestServer(t)
	rec := do(t, s, http.MethodPost, "Bearer "+testToken, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	if rec.Code != http.StatusAccepted || rec.Body.Len() != 0 {
		t.Errorf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestToolsListAndCall(t *testing.T) {
	s := newTestServer(t)
	auth := "Bearer " + testToken

	m := decodeResp(t, do(t, s, http.MethodPost, auth, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	tools := m["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 2 || tools[0].(map[string]any)["name"] != "echo" {
		t.Fatalf("tools = %v", tools)
	}
	if tools[0].(map[string]any)["inputSchema"].(map[string]any)["type"] != "object" {
		t.Error("default input schema should be an object")
	}

	m = decodeResp(t, do(t, s, http.MethodPost, auth,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"msg":"hi"}}}`))
	res := m["result"].(map[string]any)
	if res["isError"] != false {
		t.Errorf("isError = %v", res["isError"])
	}
	text := res["content"].([]any)[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, `"msg": "hi"`) {
		t.Errorf("text = %q", text)
	}

	m = decodeResp(t, do(t, s, http.MethodPost, auth,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"fail"}}`))
	res = m["result"].(map[string]any)
	if res["isError"] != true || res["content"].([]any)[0].(map[string]any)["text"] != "boom" {
		t.Errorf("tool error result = %v", res)
	}

	m = decodeResp(t, do(t, s, http.MethodPost, auth,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"nope"}}`))
	if m["error"].(map[string]any)["code"].(float64) != codeInvalidParams {
		t.Errorf("unknown tool error = %v", m["error"])
	}
}

func TestProtocolErrors(t *testing.T) {
	s := newTestServer(t)
	auth := "Bearer " + testToken

	cases := map[string]float64{
		`{"jsonrpc":"2.0","id":1,"method":"resources/list"}`: codeMethodNotFound,
		`{not json`: codeParseError,
		`{"jsonrpc":"1.0","id":1,"method":"ping"}`: codeInvalidRequest,
	}
	for body, code := range cases {
		m := decodeResp(t, do(t, s, http.MethodPost, auth, body))
		if got := m["error"].(map[string]any)["code"].(float64); got != code {
			t.Errorf("%s: code %v, want %v", body, got, code)
		}
	}

	if rec := do(t, s, http.MethodGet, auth, ""); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET: got %d", rec.Code)
	}
	if rec := do(t, s, http.MethodDelete, auth, ""); rec.Code != http.StatusNoContent {
		t.Errorf("DELETE: got %d", rec.Code)
	}
}
