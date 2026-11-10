package hopbu

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Minimal PNG signature + IHDR start; enough for http.DetectContentType.
var pngBytes = append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...)

func TestIsPublicIP(t *testing.T) {
	blocked := []string{"127.0.0.1", "10.1.2.3", "172.17.0.2", "192.168.1.1", "169.254.169.254",
		"100.64.0.1", "0.0.0.0", "::1", "fd00::1", "fe80::1", "::ffff:127.0.0.1", "64:ff9b::a00:1"}
	for _, s := range blocked {
		if isPublicIP(net.ParseIP(s)) {
			t.Errorf("%s should be blocked", s)
		}
	}
	for _, s := range []string{"8.8.8.8", "104.21.3.4", "2606:4700::1111"} {
		if !isPublicIP(net.ParseIP(s)) {
			t.Errorf("%s should be allowed", s)
		}
	}
}

func TestFetchRefusesLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(pngBytes)
	}))
	defer srv.Close()

	_, _, err := NewImageFetcher().Fetch(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "non-public") {
		t.Fatalf("want non-public refusal, got %v", err)
	}
}

func TestFetch(t *testing.T) {
	big := append(append([]byte{}, pngBytes...), bytes.Repeat([]byte{0}, MaxImageBytes)...)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok.png":
			_, _ = w.Write(pngBytes)
		case "/page":
			_, _ = w.Write([]byte("<html>not an image</html>"))
		case "/big.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(big)
		case "/redirect":
			http.Redirect(w, r, "/ok.png", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	f := newImageFetcher(func(net.IP) bool { return true })
	ctx := context.Background()

	body, ct, err := f.Fetch(ctx, srv.URL+"/ok.png")
	if err != nil || ct != "image/png" || !bytes.Equal(body, pngBytes) {
		t.Errorf("ok.png: %v %q", err, ct)
	}
	if _, ct, err := f.Fetch(ctx, srv.URL+"/redirect"); err != nil || ct != "image/png" {
		t.Errorf("redirect: %v %q", err, ct)
	}
	if _, _, err := f.Fetch(ctx, srv.URL+"/page"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("page: %v", err)
	}
	if _, _, err := f.Fetch(ctx, srv.URL+"/big.png"); err == nil || !strings.Contains(err.Error(), "2 MB") {
		t.Errorf("big: %v", err)
	}
	if _, _, err := f.Fetch(ctx, srv.URL+"/missing"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("missing: %v", err)
	}
	for _, u := range []string{"ftp://x/y.png", "/relative.png", "file:///etc/passwd"} {
		if _, _, err := f.Fetch(ctx, u); err == nil {
			t.Errorf("%s should be rejected", u)
		}
	}
}
