package mq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ErrNotRunning is the error of a client when no message queue runs for the repo.
var ErrNotRunning = errors.New("No message queue runs for this repo. Start it with: hephaestus messages serve")

// ReadDiscovery reads the discovery file of the shared directory dir.
func ReadDiscovery(dir string) (Discovery, error) {
	data, err := os.ReadFile(filepath.Join(dir, DiscoveryFile))
	if err != nil {
		return Discovery{}, err
	}
	var d Discovery
	if err := json.Unmarshal(data, &d); err != nil {
		return Discovery{}, err
	}
	return d, nil
}

// Client talks to the message queue of one repo.
type Client struct {
	discovery Discovery
	http      *http.Client
}

// Connect returns a client of the message queue whose discovery file is in the shared directory dir.
func Connect(dir string) (*Client, error) {
	d, err := ReadDiscovery(dir)
	if err != nil {
		return nil, ErrNotRunning
	}
	c := &Client{discovery: d, http: &http.Client{}}
	if c.Health() != nil {
		return nil, fmt.Errorf("The message queue at %s (pid %d) does not answer. If it stopped, start it again with: hephaestus messages serve", d.Address, d.PID)
	}
	return c, nil
}

// Health checks that the service answers.
func (c *Client) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.call(ctx, http.MethodGet, "/health", nil, nil)
}

// Send sends a message, and returns it as the service accepted it.
func (c *Client) Send(m Message) (Message, error) {
	var sent Message
	err := c.call(context.Background(), http.MethodPost, "/send", m, &sent)
	return sent, err
}

// Receive returns the messages that wait for the receiver, and waits up to r.Wait for one.
func (c *Client) Receive(ctx context.Context, r Receiver) ([]Message, error) {
	var messages []Message
	err := c.call(ctx, http.MethodPost, "/receive", r, &messages)
	return messages, err
}

// Ack acknowledges messages of the worker.
func (c *Client) Ack(worker string, ids []string) error {
	return c.call(context.Background(), http.MethodPost, "/ack", map[string]any{"worker": worker, "ids": ids}, nil)
}

// Log returns the messages after seq since.
func (c *Client) Log(since int64) ([]Message, error) {
	var messages []Message
	err := c.call(context.Background(), http.MethodGet, fmt.Sprintf("/log?since=%d", since), nil, &messages)
	return messages, err
}

func (c *Client) call(ctx context.Context, method, path string, body, result any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://"+c.discovery.Address+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.discovery.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		var failure struct{ Error string }
		if json.Unmarshal(data, &failure) == nil && failure.Error != "" {
			return errors.New(failure.Error)
		}
		return fmt.Errorf("the message queue answered %s: %s", resp.Status, bytes.TrimSpace(data))
	}
	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.Unmarshal(data, result)
}
