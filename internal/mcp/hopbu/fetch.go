package hopbu

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"
)

// MaxImageBytes matches the 2 MB limit of the browser upload endpoint.
const MaxImageBytes = 2 << 20

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// ImageFetcher downloads images from public URLs for the MCP upload tool.
//
// Because the server fetches a URL chosen by the caller, it refuses to connect
// to loopback, private, link-local and other non-public addresses (SSRF
// guard). The check runs on the IP actually dialled, so DNS tricks and
// redirects to internal hosts are refused too.
type ImageFetcher struct {
	client *http.Client
}

// NewImageFetcher returns a fetcher that only reaches public IPs.
func NewImageFetcher() *ImageFetcher {
	return newImageFetcher(isPublicIP)
}

func newImageFetcher(allowIP func(net.IP) bool) *ImageFetcher {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip == nil || !allowIP(ip) {
				return fmt.Errorf("refusing to connect to non-public address %s", host)
			}
			return nil
		},
	}
	transport := &http.Transport{
		Proxy:                 nil, // dial targets directly so the IP check applies
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		MaxIdleConns:          4,
		IdleConnTimeout:       30 * time.Second,
	}
	return &ImageFetcher{client: &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("redirect to unsupported scheme")
			}
			return nil
		},
	}}
}

// Fetch downloads rawURL and returns its bytes and sniffed content type.
// Only JPEG, PNG and WebP up to MaxImageBytes are accepted.
func (f *ImageFetcher) Fetch(ctx context.Context, rawURL string) ([]byte, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, "", errors.New("image_url must be an absolute http(s) URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("build request: %w", err)
	}
	// Some image hosts refuse requests without a browser-like user agent.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; HopbuBot/1.0; +https://hopbu)")
	req.Header.Set("Accept", "image/webp,image/png,image/jpeg;q=0.9,*/*;q=0.5")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download image: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > MaxImageBytes {
		return nil, "", fmt.Errorf("image is %d bytes; the limit is 2 MB", resp.ContentLength)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxImageBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("read image: %w", err)
	}
	if len(body) > MaxImageBytes {
		return nil, "", errors.New("image is larger than the 2 MB limit")
	}

	contentType := http.DetectContentType(body)
	if !allowedImageTypes[contentType] {
		return nil, "", fmt.Errorf("unsupported image type %q; use JPEG, PNG or WebP", contentType)
	}
	return body, contentType, nil
}

var nonPublicNets = mustCIDRs(
	"0.0.0.0/8",     // "this" network
	"100.64.0.0/10", // carrier-grade NAT
	"192.0.0.0/24",  // IETF protocol assignments
	"192.0.2.0/24",  // documentation
	"198.18.0.0/15", // benchmarking
	"198.51.100.0/24",
	"203.0.113.0/24",
	"240.0.0.0/4",  // reserved
	"64:ff9b::/96", // NAT64 can map to internal IPv4
	"2001:db8::/32",
)

func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return false
	}
	for _, n := range nonPublicNets {
		if n.Contains(ip) {
			return false
		}
	}
	return true
}

func mustCIDRs(cidrs ...string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			panic(err)
		}
		out = append(out, n)
	}
	return out
}
