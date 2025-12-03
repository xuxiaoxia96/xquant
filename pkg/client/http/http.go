package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

// Client HTTP客户端封装
type Client struct {
	client *http.Client
}

// Config HTTP客户端配置
type Config struct {
	// Timeout 请求超时时间，默认30秒
	Timeout time.Duration
	// MaxIdleConns 最大空闲连接数，默认100
	MaxIdleConns int
	// MaxIdleConnsPerHost 每个主机的最大空闲连接数，默认10
	MaxIdleConnsPerHost int
	// MaxConnsPerHost 每个主机的最大连接数，默认0（无限制）
	MaxConnsPerHost int
	// IdleConnTimeout 空闲连接超时时间，默认90秒
	IdleConnTimeout time.Duration
	// DisableKeepAlives 是否禁用keep-alive，默认false
	DisableKeepAlives bool
	// TLSHandshakeTimeout TLS握手超时时间，默认10秒
	TLSHandshakeTimeout time.Duration
	// ResponseHeaderTimeout 响应头超时时间，默认0（无限制）
	ResponseHeaderTimeout time.Duration
	// ExpectContinueTimeout Expect: 100-continue超时时间，默认1秒
	ExpectContinueTimeout time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:               30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       0, // 0表示无限制
		IdleConnTimeout:       90 * time.Second,
		DisableKeepAlives:     false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 0, // 0表示无限制
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewClient 创建新的HTTP客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	transport := &http.Transport{
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &Client{
		client: client,
	}
}

// HttpClient 默认HTTP客户端实例（单例）
var HttpClient *Client

// init 初始化默认HTTP客户端
func init() {
	HttpClient = NewClient(DefaultConfig())
}

// Do 执行HTTP请求
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// DoWithContext 使用context执行HTTP请求
func (c *Client) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return c.client.Do(req)
}

// Get 发送GET请求
func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

// GetWithContext 使用context发送GET请求
func (c *Client) GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

// Post 发送POST请求
func (c *Client) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.client.Do(req)
}

// PostWithContext 使用context发送POST请求
func (c *Client) PostWithContext(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.client.Do(req)
}

// PostJSON 发送JSON格式的POST请求
func (c *Client) PostJSON(url string, body []byte) (*http.Response, error) {
	return c.Post(url, "application/json", bytes.NewReader(body))
}

// PostJSONWithContext 使用context发送JSON格式的POST请求
func (c *Client) PostJSONWithContext(ctx context.Context, url string, body []byte) (*http.Response, error) {
	return c.PostWithContext(ctx, url, "application/json", bytes.NewReader(body))
}

// PostForm 发送表单格式的POST请求
func (c *Client) PostForm(url string, body io.Reader) (*http.Response, error) {
	return c.Post(url, "application/x-www-form-urlencoded", body)
}

// PostFormWithContext 使用context发送表单格式的POST请求
func (c *Client) PostFormWithContext(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	return c.PostWithContext(ctx, url, "application/x-www-form-urlencoded", body)
}

// GetBody 发送GET请求并读取响应体
func (c *Client) GetBody(url string) ([]byte, error) {
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// GetBodyWithContext 使用context发送GET请求并读取响应体
func (c *Client) GetBodyWithContext(ctx context.Context, url string) ([]byte, error) {
	resp, err := c.GetWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// PostBody 发送POST请求并读取响应体
func (c *Client) PostBody(url string, contentType string, body io.Reader) ([]byte, error) {
	resp, err := c.Post(url, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// PostBodyWithContext 使用context发送POST请求并读取响应体
func (c *Client) PostBodyWithContext(ctx context.Context, url string, contentType string, body io.Reader) ([]byte, error) {
	resp, err := c.PostWithContext(ctx, url, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// GetClient 获取底层的http.Client（用于高级用法）
func (c *Client) GetClient() *http.Client {
	return c.client
}

// 以下为包级别的便捷函数，使用DefaultClient

// Get 使用默认客户端发送GET请求
func Get(url string) (*http.Response, error) {
	return HttpClient.Get(url)
}

// GetWithContext 使用默认客户端和context发送GET请求
func GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	return HttpClient.GetWithContext(ctx, url)
}

// Post 使用默认客户端发送POST请求
func Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	return HttpClient.Post(url, contentType, body)
}

// PostWithContext 使用默认客户端和context发送POST请求
func PostWithContext(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	return HttpClient.PostWithContext(ctx, url, contentType, body)
}

// PostJSON 使用默认客户端发送JSON格式的POST请求
func PostJSON(url string, body []byte) (*http.Response, error) {
	return HttpClient.PostJSON(url, body)
}

// PostJSONWithContext 使用默认客户端和context发送JSON格式的POST请求
func PostJSONWithContext(ctx context.Context, url string, body []byte) (*http.Response, error) {
	return HttpClient.PostJSONWithContext(ctx, url, body)
}

// PostForm 使用默认客户端发送表单格式的POST请求
func PostForm(url string, body io.Reader) (*http.Response, error) {
	return HttpClient.PostForm(url, body)
}

// PostFormWithContext 使用默认客户端和context发送表单格式的POST请求
func PostFormWithContext(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	return HttpClient.PostFormWithContext(ctx, url, body)
}

// PostString 使用默认客户端发送POST请求（body为字符串，返回响应体）
func PostString(url string, body string) ([]byte, error) {
	var reader io.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	}
	return HttpClient.PostBody(url, "application/x-www-form-urlencoded", reader)
}
