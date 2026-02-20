package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:132.0) Gecko/20100101 Firefox/132.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0",
}

// Options configures the anti-ban HTTP client.
type Options struct {
	ProxyURL   string
	MinDelay   time.Duration
	MaxDelay   time.Duration
	MaxRetries int
}

func (o Options) withDefaults() Options {
	if o.MinDelay == 0 {
		o.MinDelay = 2 * time.Second
	}
	if o.MaxDelay == 0 {
		o.MaxDelay = 5 * time.Second
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = 3
	}
	return o
}

// Client wraps http.Client with anti-ban protections.
type Client struct {
	inner      *http.Client
	mu         sync.Mutex
	lastReq    map[string]time.Time
	minDelay   time.Duration
	maxDelay   time.Duration
	maxRetries int
}

// New creates a Client with the given options.
func New(opts Options) (*Client, error) {
	opts = opts.withDefaults()

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	if opts.ProxyURL != "" {
		proxyURL, err := url.Parse(opts.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("httpclient: invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &Client{
		inner:      &http.Client{Transport: transport},
		lastReq:    make(map[string]time.Time),
		minDelay:   opts.MinDelay,
		maxDelay:   opts.MaxDelay,
		maxRetries: opts.MaxRetries,
	}, nil
}

// Do executes the request with UA rotation, realistic headers, rate limiting,
// and retry with exponential backoff.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.setHeaders(req)

	if err := c.rateLimit(req.Context(), req.URL.Host); err != nil {
		return nil, err
	}

	var resp *http.Response
	var err error

	for attempt := range c.maxRetries {
		resp, err = c.inner.Do(req)
		if err != nil {
			return nil, fmt.Errorf("httpclient: request failed: %w", err)
		}

		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
			return resp, nil
		}

		resp.Body.Close()
		backoff := time.Duration(1<<uint(attempt)) * 2 * time.Second
		fmt.Printf("[httpclient] %s retornou %d, aguardando %v (tentativa %d/%d)\n",
			req.URL.Host, resp.StatusCode, backoff, attempt+1, c.maxRetries)

		select {
		case <-time.After(backoff):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}

	return resp, err
}

func (c *Client) setHeaders(req *http.Request) {
	ua := userAgents[rand.Intn(len(userAgents))]
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	// Accept-Encoding é omitido intencionalmente: Go's http.Transport
	// gerencia compressão automaticamente quando este header não é setado.
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Connection", "keep-alive")
}

func (c *Client) rateLimit(ctx context.Context, host string) error {
	c.mu.Lock()
	last, ok := c.lastReq[host]
	c.lastReq[host] = time.Now()
	c.mu.Unlock()

	if !ok {
		return nil
	}

	elapsed := time.Since(last)
	delay := c.minDelay + time.Duration(rand.Int63n(int64(c.maxDelay-c.minDelay)))

	if elapsed < delay {
		wait := delay - elapsed
		fmt.Printf("[httpclient] rate limit: aguardando %v antes de acessar %s\n", wait.Round(time.Millisecond), host)
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	c.mu.Lock()
	c.lastReq[host] = time.Now()
	c.mu.Unlock()

	return nil
}
