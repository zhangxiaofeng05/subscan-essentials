package chainrpc

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/zhangxiaofeng05/subscan/internal/utils/bigint"
)

func (c *Client) ChainGetBlockHash(ctx context.Context, height int64) (string, error) {
	var hash string
	err := c.jsonrpc.CallContext(ctx, &hash, "chain_getBlockHash", height)
	if err != nil {
		return "", err
	}
	if hash == "" {
		return "", fmt.Errorf("chain_getBlockHash height:%v hash is nil", height)
	}
	return hash, nil
}

type ChainNewHeadLog struct {
	Logs []string `json:"logs"`
}

type ChHeader struct {
	ParentHash     string          `json:"parentHash"`
	Number         bigint.Int      `json:"number"`
	StateRoot      string          `json:"stateRoot"`
	ExtrinsicsRoot string          `json:"extrinsicsRoot"`
	Digest         ChainNewHeadLog `json:"digest"`
}

type ChBlock struct {
	Header     ChHeader `json:"header"`
	Extrinsics []string `json:"extrinsics"`
}

type ChainBlock struct {
	Block ChBlock `json:"block"`
}

func (c *Client) ChainGetBlock(ctx context.Context, hash string) (*ChainBlock, error) {
	var res ChainBlock
	err := c.jsonrpc.CallContext(ctx, &res, "chain_getBlock", hash)
	if err != nil {
		return nil, fmt.Errorf("chain_getBlock hash:%v err:%v", hash, err)
	}
	return &res, nil
}

func (c *Client) StateGetMetadata(ctx context.Context, blockHash string) (string, error) {
	var res string
	err := c.jsonrpc.CallContext(ctx, &res, "state_getMetadata", blockHash)
	if err != nil {
		return "", fmt.Errorf("state_getMetadata hash:%v err:%v", blockHash, err)
	}
	return res, nil
}

func (c *Client) StateGetStorageAt(ctx context.Context, storageKey, blockHash string) (string, error) {
	var res string
	err := c.jsonrpc.CallContext(ctx, &res, "state_getStorageAt", storageKey, blockHash)
	if err != nil {
		return "", fmt.Errorf("state_getStorageAt hash:%v err:%v", blockHash, err)
	}
	return res, nil
}

type RuntimeVersion struct {
	Apis             [][]interface{} `json:"apis"`
	AuthoringVersion int             `json:"authoringVersion"`
	ImplName         string          `json:"implName"`
	ImplVersion      int             `json:"implVersion"`
	SpecName         string          `json:"specName"`
	SpecVersion      int             `json:"specVersion"`
}

func (c *Client) ChainGetRuntimeVersion(ctx context.Context, blockHash string) (*RuntimeVersion, error) {
	var res RuntimeVersion
	err := c.jsonrpc.CallContext(ctx, &res, "chain_getRuntimeVersion", blockHash)
	if err != nil {
		return nil, fmt.Errorf("chain_getRuntimeVersion hash:%v err:%v", blockHash, err)
	}
	return &res, nil
}

type PaymentQueryInfo struct {
	Class      string          `json:"class"`
	PartialFee decimal.Decimal `json:"partialFee"`
	Weight     int64           `json:"weight"`
}

func (c *Client) PaymentQueryInfo(ctx context.Context, encodedExtrinsic string) (*PaymentQueryInfo, error) {
	var res PaymentQueryInfo
	err := c.jsonrpc.CallContext(ctx, &res, "payment_queryInfo", encodedExtrinsic)
	if err != nil {
		return nil, fmt.Errorf("payment_queryInfo err:%v", err)
	}
	return &res, nil
}
