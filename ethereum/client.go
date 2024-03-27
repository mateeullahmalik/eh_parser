package ethereum

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/mateeullahmalik/eh_parser/ethereum/jsonrpc"
)

type GetTransactionsResult []GetTransactionResult

type GetTransactionResult struct {
	Amount float64 `json:"amount"`
	Fee    float64 `json:"fee"`
	Block  int32   `json:"block"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	TxID   string  `json:"txid"`
}

type client struct {
	jsonrpc.RPCClient
}

func (client *client) GetBlockCount(ctx context.Context) (int32, error) {
	res, err := client.CallWithContext(ctx, "getblockcount")
	if err != nil {
		return 0, fmt.Errorf("failed to call getblockcount: %w", err)
	}

	if res.Error != nil {
		return 0, fmt.Errorf("failed to get block count: %w", res.Error)
	}

	cnt, err := res.GetInt()

	return int32(cnt), err
}

func (client *client) GetTransaction(ctx context.Context, txID string) (GetTransactionResult, error) {
	result := GetTransactionResult{}

	if err := client.callFor(ctx, &result, "gettransaction", txID); err != nil {
		return result, fmt.Errorf("failed to get transaction: %w", err)
	}

	return result, nil
}

func (client *client) GetTransactionsByAddress(ctx context.Context, address string, block int32) (GetTransactionsResult, error) {
	result := GetTransactionsResult{}

	if err := client.callFor(ctx, &result, "transaction", "list", block); err != nil {
		return result, fmt.Errorf("failed to get transaction: %w", err)
	}

	return result, nil
}

func (client *client) callFor(ctx context.Context, object interface{}, method string, params ...interface{}) error {
	return client.CallForWithContext(ctx, object, method, params)
}

// NewClient returns a new Client instance.
func NewClient(config *Config) *client {
	//Configure network addressing
	endpoint := net.JoinHostPort(config.Hostname, strconv.Itoa(config.Port))
	if !strings.Contains(endpoint, "//") {
		endpoint = "http://" + endpoint
	}

	//Parse and configure RPC authorization headers
	opts := &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(config.Username+":"+config.Password)),
		},
	}

	//Return a Client interface with the proper RPCClient configurations
	return &client{
		RPCClient: jsonrpc.NewClientWithOpts(endpoint, opts),
	}
}
