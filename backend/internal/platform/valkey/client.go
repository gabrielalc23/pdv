package valkey

import (
	"context"
	"fmt"
	"net"
	"time"

	valkey "github.com/valkey-io/valkey-go"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

func (c Config) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("valkey address is required")
	}
	if c.DB < 0 || c.DB > 15 {
		return fmt.Errorf("valkey db must be between 0 and 15")
	}
	return nil
}

type Client struct {
	client valkey.Client
	cfg    Config
}

func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	opts := valkey.ClientOption{
		InitAddress: []string{cfg.Addr},
		Password:    cfg.Password,
		SelectDB:    cfg.DB,
		Dialer: net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 10 * time.Second,
		},
		ConnWriteTimeout: 5 * time.Second,
	}

	client, err := valkey.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}

	return &Client{client: client, cfg: cfg}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	result := c.client.Do(ctx, c.client.B().Ping().Build())
	if err := result.Error(); err != nil {
		return fmt.Errorf("valkey ping failed: %w", err)
	}
	return nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.Ping(ctx)
}

func (c *Client) Do(ctx context.Context, cmd valkey.Completed) (valkey.ValkeyResult, error) {
	result := c.client.Do(ctx, cmd)
	return result, result.Error()
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) B() valkey.Builder {
	return c.client.B()
}
