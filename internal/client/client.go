package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/okonomipizza/statecast/internal/runtime"
	"github.com/okonomipizza/statecast/internal/store"
)

const requestTimeout = 5 * time.Second

// Client は daemon へ HTTP over Unix socket で接続する。
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// New は Unix ソケット経由の daemon クライアントを返す。
func New() (*Client, error) {
	socketPath, err := runtime.SocketPath()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   requestTimeout,
		},
		baseURL: "http://statecast",
	}, nil
}

type registerRequest struct {
	Name string `json:"name"`
}

// Register はエージェントを登録する。
func (c *Client) Register(name string) (store.Agent, error) {
	body, err := json.Marshal(registerRequest{Name: name})
	if err != nil {
		return store.Agent{}, fmt.Errorf("encode register request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/v1/agents", "application/json", bytes.NewReader(body))
	if err != nil {
		return store.Agent{}, fmt.Errorf("daemon is not running")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return store.Agent{}, readError(resp)
	}

	var agent store.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		return store.Agent{}, fmt.Errorf("decode register response: %w", err)
	}

	return agent, nil
}

// List は登録済みエージェント一覧を取得する。
func (c *Client) List() ([]store.Agent, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/v1/agents")
	if err != nil {
		return nil, fmt.Errorf("daemon is not running")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var agents []store.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}

	return agents, nil
}

// GetState は指定エージェントの状態 JSON を取得する。
func (c *Client) GetState(name string) (json.RawMessage, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/v1/agents/" + url.PathEscape(name) + "/state")
	if err != nil {
		return nil, fmt.Errorf("daemon is not running")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("agent not found: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read get response: %w", err)
	}

	return body, nil
}

// UpdateState は指定エージェントの状態を更新する。
func (c *Client) UpdateState(name string, state json.RawMessage) error {
	req, err := http.NewRequest(http.MethodPut, c.baseURL+"/v1/agents/"+url.PathEscape(name)+"/state", bytes.NewReader(state))
	if err != nil {
		return fmt.Errorf("create update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("daemon is not running")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("agent not found: %s", name)
	}
	if resp.StatusCode != http.StatusNoContent {
		return readError(resp)
	}

	return nil
}

// Ping は daemon が応答可能か確認する。
func (c *Client) Ping() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("daemon is not running")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return readError(resp)
	}

	return nil
}

func readError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	if len(body) == 0 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	return fmt.Errorf("%s", bytes.TrimSpace(body))
}
