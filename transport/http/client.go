package http

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/errors"
)

// ClientOption is HTTP client option.
type ClientOption func(*Client)

// ClientDecodeErrorFunc is client error decoder.
type ClientDecodeErrorFunc func(res *http.Response) error

// ClientDialTimeout with client dial timeout.
func ClientDialTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.dialTimeout = timeout
	}
}

// ClientTimeout with client request timeout.
func ClientTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// ClientKeepAlive with client keepavlie.
func ClientKeepAlive(ka time.Duration) ClientOption {
	return func(c *Client) {
		c.keepAlive = ka
	}
}

// ClientUserAgent with client user agent.
func ClientUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// ClientErrorDecoder with client error decoder.
func ClientErrorDecoder(d ClientDecodeErrorFunc) ClientOption {
	return func(c *Client) {
		c.errorDecoder = d
	}
}

// Client is a HTTP transport client.
type Client struct {
	base         http.RoundTripper
	dialTimeout  time.Duration
	timeout      time.Duration
	keepAlive    time.Duration
	userAgent    string
	errorDecoder ClientDecodeErrorFunc
}

// NewClient new a HTTP transport client.
func NewClient(opts ...ClientOption) (*http.Client, error) {
	client := &Client{
		dialTimeout:  200 * time.Millisecond,
		timeout:      500 * time.Millisecond,
		keepAlive:    30 * time.Second,
		errorDecoder: ClientDecodeError,
	}
	for _, o := range opts {
		o(client)
	}
	client.base = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   client.dialTimeout,
			KeepAlive: client.keepAlive,
		}).DialContext,
	}
	return &http.Client{
		Transport: client,
	}, nil
}

// RoundTrip is transport round trip.
func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	ctx, cancel := context.WithTimeout(req.Context(), c.timeout)
	defer cancel()
	res, err := c.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if err := c.errorDecoder(res); err != nil {
		return nil, err
	}
	return res, nil
}

// ClientDecodeError is default error decoder.
func ClientDecodeError(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	defer res.Body.Close()
	slurp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	contentType := res.Header.Get("content-type")
	codec := encoding.GetCodec(contentSubtype(contentType))
	if codec == nil {
		return errors.Internal("UnknownCodec", contentType)
	}
	se := &errors.StatusError{}
	if err := codec.Unmarshal(slurp, se); err != nil {
		return err
	}
	return se
}

// ClientDecodeBody decode response body.
func ClientDecodeBody(res *http.Response, v interface{}) error {
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	contentType := res.Header.Get("content-type")
	codec := encoding.GetCodec(contentSubtype(contentType))
	if codec == nil {
		return errors.Internal("UnknownCodec", contentType)
	}
	return codec.Unmarshal(data, v)
}
