package common

import (
	"net"
	"net/http"
	"time"
)

// HttpClient is a simple http client with timeout.
type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	transport := &http.Transport{}
	transport.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))
		return conn, nil
	}
	client := &http.Client{
		Transport: transport,
	}
	return &HTTPClient{
		client: client,
	}
}

func (c *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}
