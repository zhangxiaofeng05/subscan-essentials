package chainrpc

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/rpc"
	"time"
)

type JSON interface {
	Call(result interface{}, method string, args ...interface{}) error
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
	Close()
}

type Client struct {
	jsonrpc JSON
}

func (c *Client) Close() {
	c.jsonrpc.Close()
}

func DialForWallet(rawurl string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return DialContextForWallet(ctx, rawurl)
}

func DialContextForWallet(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return &Client{
		jsonrpc: c,
	}, nil
}

type Response struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}
